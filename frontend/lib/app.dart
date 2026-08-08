import 'package:flutter/material.dart';

import 'core/app_constants.dart';
import 'features/auth/login_page.dart';
import 'features/home/home_shell.dart';
import 'models.dart';
import 'services/api_client.dart';
import 'services/blob_search_registry.dart';
import 'services/session_store.dart';
import 'theme/app_theme.dart';

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return const OdessaApp();
  }
}

class OdessaApp extends StatefulWidget {
  const OdessaApp({super.key});

  @override
  State<OdessaApp> createState() => _OdessaAppState();
}

class _OdessaAppState extends State<OdessaApp> {
  late String _apiBaseUrl;
  late ApiClient _apiClient;
  final SessionStore _sessionStore = SessionStore();
  bool _loadingSession = true;
  AuthSession? _session;

  @override
  void initState() {
    super.initState();
    _apiBaseUrl = AppConstants.defaultApiUrl;
    _apiClient = ApiClient(baseUrl: _apiBaseUrl);
    _initializeApp();
  }

  Future<void> _initializeApp() async {
    await BlobSearchRegistry.instance.load();
    await _restoreSession();
  }

  Future<void> _restoreSession() async {
    final stored = await _sessionStore.load();
    if (stored == null) {
      if (!mounted) {
        return;
      }
      setState(() {
        _loadingSession = false;
      });
      return;
    }

    try {
      final refreshed = await _apiClient.refreshSession(stored.refreshToken);
      await _sessionStore.save(refreshed);

      if (!mounted) {
        return;
      }
      setState(() {
        _session = refreshed;
        _loadingSession = false;
      });
    } catch (_) {
      await _sessionStore.clear();
      if (!mounted) {
        return;
      }
      setState(() {
        _session = null;
        _loadingSession = false;
      });
    }
  }

  void _setApiBaseUrl(String nextBaseUrl) {
    setState(() {
      _apiBaseUrl = nextBaseUrl;
      _apiClient = ApiClient(baseUrl: nextBaseUrl);
    });
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      debugShowCheckedModeBanner: false,
      title: 'Odessa Catalog',
      theme: OdessaTheme.dark(),
      home: SelectionArea(
        child: _loadingSession
            ? const Scaffold(body: Center(child: CircularProgressIndicator()))
            : _session == null
            ? LoginPage(
                apiClient: _apiClient,
                onLoginSuccess: (session) async {
                  await _sessionStore.save(session);
                  if (!mounted) {
                    return;
                  }
                  setState(() {
                    _session = session;
                  });
                },
              )
            : HomeShell(
                apiClient: _apiClient,
                session: _session!,
                currentApiUrl: _apiBaseUrl,
                onApiUrlChanged: _setApiBaseUrl,
                onSessionChanged: (session) async {
                  await _sessionStore.save(session);
                  if (!mounted) {
                    return;
                  }
                  setState(() {
                    _session = session;
                  });
                },
                onLogout: () async {
                  await _sessionStore.clear();
                  if (!mounted) {
                    return;
                  }
                  setState(() {
                    _session = null;
                  });
                },
              ),
      ),
    );
  }
}
