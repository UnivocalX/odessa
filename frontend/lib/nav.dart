import 'package:flutter/material.dart';

import 'features/home/home_shell.dart';
import 'models.dart';
import 'services/api_client.dart';

class NavRail extends StatefulWidget {
  const NavRail({
    super.key,
    required this.apiClient,
    required this.session,
    required this.currentApiUrl,
    required this.onApiUrlChanged,
    required this.onSessionChanged,
    required this.onLogout,
  });

  final ApiClient apiClient;
  final AuthSession session;
  final String currentApiUrl;
  final ValueChanged<String> onApiUrlChanged;
  final ValueChanged<AuthSession> onSessionChanged;
  final VoidCallback onLogout;

  @override
  State<NavRail> createState() => _NavRailState();
}

class _NavRailState extends State<NavRail> {
  @override
  Widget build(BuildContext context) {
    return HomeShell(
      apiClient: widget.apiClient,
      session: widget.session,
      currentApiUrl: widget.currentApiUrl,
      onApiUrlChanged: widget.onApiUrlChanged,
      onSessionChanged: widget.onSessionChanged,
      onLogout: widget.onLogout,
    );
  }
}
