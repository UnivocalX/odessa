import 'package:flutter/material.dart';

import '../../colors.dart';
import '../../core/app_constants.dart';
import '../../models.dart';
import '../../services/api_client.dart';
import '../../services/storage_download_config_store.dart';

class SettingsView extends StatefulWidget {
  const SettingsView({
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
  State<SettingsView> createState() => _SettingsViewState();
}

class _SettingsViewState extends State<SettingsView> {
  late final TextEditingController _apiUrlController;
  final TextEditingController _storageTemplateController =
      TextEditingController();
  final TextEditingController _storageS3EndpointController =
      TextEditingController();
  final TextEditingController _storageBearerTokenController =
      TextEditingController();
  final TextEditingController _storageAccessKeyController =
      TextEditingController();
  final TextEditingController _storageSecretKeyController =
      TextEditingController();
  final TextEditingController _storageHeaderNameController =
      TextEditingController();
  final TextEditingController _storageHeaderValueController =
      TextEditingController();
  final TextEditingController _currentPasswordController =
      TextEditingController();
  final TextEditingController _newPasswordController = TextEditingController();
  final TextEditingController _disablePasswordController =
      TextEditingController();

  bool _busy = false;
  final StorageDownloadConfigStore _storageConfigStore =
      StorageDownloadConfigStore();

  @override
  void initState() {
    super.initState();
    _apiUrlController = TextEditingController(text: widget.currentApiUrl);
    _loadStorageConfig();
  }

  Future<void> _loadStorageConfig() async {
    final config = await _storageConfigStore.load();
    if (!mounted) {
      return;
    }
    setState(() {
      _storageTemplateController.text = config.downloadUrlTemplate;
      _storageS3EndpointController.text = config.s3ApiEndpoint;
      _storageBearerTokenController.text = config.bearerToken;
      _storageAccessKeyController.text = config.accessKey;
      _storageSecretKeyController.text = config.secretKey;
      _storageHeaderNameController.text = config.customHeaderName;
      _storageHeaderValueController.text = config.customHeaderValue;
    });
  }

  @override
  void didUpdateWidget(covariant SettingsView oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.currentApiUrl != widget.currentApiUrl) {
      _apiUrlController.text = widget.currentApiUrl;
    }
  }

  @override
  void dispose() {
    _apiUrlController.dispose();
    _storageTemplateController.dispose();
    _storageS3EndpointController.dispose();
    _storageBearerTokenController.dispose();
    _storageAccessKeyController.dispose();
    _storageSecretKeyController.dispose();
    _storageHeaderNameController.dispose();
    _storageHeaderValueController.dispose();
    _currentPasswordController.dispose();
    _newPasswordController.dispose();
    _disablePasswordController.dispose();
    super.dispose();
  }

  void _showMessage(String message) {
    if (!mounted) {
      return;
    }
    ScaffoldMessenger.of(
      context,
    ).showSnackBar(SnackBar(content: Text(message)));
  }

  Future<void> _runBusy(Future<void> Function() action) async {
    setState(() {
      _busy = true;
    });
    try {
      await action();
    } finally {
      if (mounted) {
        setState(() {
          _busy = false;
        });
      }
    }
  }

  bool _isValidPassword(String value) {
    return value.length >= AppConstants.minPasswordLength;
  }

  Widget _sectionCard({required Widget child}) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(AppConstants.cardPadding),
        child: child,
      ),
    );
  }

  Future<void> _refreshSession() async {
    await _runBusy(() async {
      try {
        final session = await widget.apiClient.refreshSession(
          widget.session.refreshToken,
        );
        widget.onSessionChanged(session);
        _showMessage('Session refreshed');
      } catch (error) {
        _showMessage(error.toString());
      }
    });
  }

  Future<void> _saveApiUrl() async {
    final text = _apiUrlController.text.trim();
    if (text.isEmpty) {
      _showMessage('API URL cannot be empty');
      return;
    }

    try {
      final uri = Uri.parse(text);
      if (!uri.hasScheme || uri.host.isEmpty) {
        throw const FormatException('Invalid API URL');
      }
      widget.onApiUrlChanged(text);
      _showMessage('API URL updated');
    } catch (_) {
      _showMessage('Please enter a valid API URL');
    }
  }

  Future<void> _saveStorageConfig() async {
    final config = StorageDownloadConfig(
      downloadUrlTemplate: _storageTemplateController.text.trim(),
      s3ApiEndpoint: _storageS3EndpointController.text.trim(),
      bearerToken: _storageBearerTokenController.text.trim(),
      accessKey: _storageAccessKeyController.text.trim(),
      secretKey: _storageSecretKeyController.text.trim(),
      customHeaderName: _storageHeaderNameController.text.trim(),
      customHeaderValue: _storageHeaderValueController.text.trim(),
    );
    await _storageConfigStore.save(config);
    _showMessage('Storage download configuration saved');
  }

  Future<void> _changePassword() async {
    final current = _currentPasswordController.text;
    final next = _newPasswordController.text;
    if (!_isValidPassword(current) || !_isValidPassword(next)) {
      _showMessage(
        'Passwords must be at least ${AppConstants.minPasswordLength} characters',
      );
      return;
    }

    await _runBusy(() async {
      try {
        final message = await widget.apiClient.changePassword(
          accessToken: widget.session.accessToken,
          currentPassword: current,
          newPassword: next,
        );
        _currentPasswordController.clear();
        _newPasswordController.clear();
        _showMessage(message);
      } catch (error) {
        _showMessage(error.toString());
      }
    });
  }

  Future<void> _logout() async {
    await _runBusy(() async {
      try {
        await widget.apiClient.logout(
          accessToken: widget.session.accessToken,
          refreshToken: widget.session.refreshToken,
        );
      } catch (_) {
        // Logout on backend failure should not block local logout.
      }
      widget.onLogout();
    });
  }

  Future<void> _disableAccount() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) {
        return AlertDialog(
          title: const Text('Disable account'),
          content: const Text(
            'This will disable your account. You will be logged out immediately.',
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(context).pop(false),
              child: const Text('Cancel'),
            ),
            FilledButton(
              onPressed: () => Navigator.of(context).pop(true),
              child: const Text('Disable'),
            ),
          ],
        );
      },
    );

    if (confirmed != true) {
      return;
    }

    final password = _disablePasswordController.text;
    if (!_isValidPassword(password)) {
      _showMessage(
        'Password must be at least ${AppConstants.minPasswordLength} characters',
      );
      return;
    }

    await _runBusy(() async {
      try {
        await widget.apiClient.disableAccount(
          accessToken: widget.session.accessToken,
          password: password,
        );
        widget.onLogout();
      } catch (error) {
        _showMessage(error.toString());
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    return ListView(
      children: [
        const ListTile(
          leading: Icon(Icons.person, color: OdessaColors.accentDark),
          title: Text('User Information'),
          subtitle: Text('Session and account controls'),
        ),
        _sectionCard(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text(
                'Session',
                style: TextStyle(fontWeight: FontWeight.w700),
              ),
              const SizedBox(height: 8),
              Text('Token type: ${widget.session.tokenType}'),
              Text('Access expires in: ${widget.session.expiresIn} sec'),
              const SizedBox(height: 12),
              Wrap(
                spacing: 8,
                runSpacing: 8,
                children: [
                  FilledButton.icon(
                    onPressed: _busy ? null : _refreshSession,
                    icon: const Icon(Icons.refresh),
                    label: const Text('Refresh session'),
                  ),
                  OutlinedButton.icon(
                    onPressed: _busy ? null : _logout,
                    icon: const Icon(Icons.logout),
                    label: const Text('Logout'),
                  ),
                ],
              ),
            ],
          ),
        ),
        const SizedBox(height: AppConstants.sectionSpacing),
        _sectionCard(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text(
                'Storage Download APIs & Keys',
                style: TextStyle(fontWeight: FontWeight.w700),
              ),
              const SizedBox(height: 8),
              const Text(
                'Used for direct storage downloads. Public HTTP URLs use the template below, and S3/MinIO downloads use the endpoint and credentials below to sign requests directly from the client.',
              ),
              const SizedBox(height: 10),
              TextField(
                controller: _storageTemplateController,
                decoration: const InputDecoration(
                  labelText: 'Public download URL template',
                  hintText:
                      'https://storage.example.com/download?uri={location}&hash={hash}',
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 10),
              const Text(
                'S3 section',
                style: TextStyle(fontWeight: FontWeight.w700),
              ),
              const SizedBox(height: 8),
              TextField(
                controller: _storageS3EndpointController,
                decoration: const InputDecoration(
                  labelText: 'S3 / MinIO endpoint (optional)',
                  hintText:
                      'http://localhost:9000',
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 10),
              TextField(
                controller: _storageBearerTokenController,
                obscureText: true,
                decoration: const InputDecoration(
                  labelText: 'Bearer token (optional)',
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 10),
              TextField(
                controller: _storageAccessKeyController,
                decoration: const InputDecoration(
                  labelText: 'Access key (required for S3 / MinIO)',
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 10),
              TextField(
                controller: _storageSecretKeyController,
                obscureText: true,
                decoration: const InputDecoration(
                  labelText: 'Secret key (required for S3 / MinIO)',
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 10),
              Row(
                children: [
                  Expanded(
                    child: TextField(
                      controller: _storageHeaderNameController,
                      decoration: const InputDecoration(
                        labelText: 'Custom header name',
                        border: OutlineInputBorder(),
                      ),
                    ),
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    child: TextField(
                      controller: _storageHeaderValueController,
                      decoration: const InputDecoration(
                        labelText: 'Custom header value',
                        border: OutlineInputBorder(),
                      ),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 12),
              FilledButton.icon(
                onPressed: _busy ? null : _saveStorageConfig,
                icon: const Icon(Icons.save_outlined),
                label: const Text('Save storage configuration'),
              ),
            ],
          ),
        ),
        const SizedBox(height: AppConstants.sectionSpacing),
        _sectionCard(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text(
                'API Endpoint',
                style: TextStyle(fontWeight: FontWeight.w700),
              ),
              const SizedBox(height: 8),
              TextField(
                controller: _apiUrlController,
                decoration: const InputDecoration(
                  labelText: 'API Base URL',
                  hintText: 'http://localhost:9090',
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 12),
              FilledButton(
                onPressed: _busy ? null : _saveApiUrl,
                child: const Text('Save API URL'),
              ),
            ],
          ),
        ),
        const SizedBox(height: AppConstants.sectionSpacing),
        _sectionCard(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text(
                'Change Password',
                style: TextStyle(fontWeight: FontWeight.w700),
              ),
              const SizedBox(height: 8),
              TextField(
                controller: _currentPasswordController,
                obscureText: true,
                decoration: const InputDecoration(
                  labelText: 'Current password',
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 10),
              TextField(
                controller: _newPasswordController,
                obscureText: true,
                decoration: const InputDecoration(
                  labelText: 'New password',
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 12),
              FilledButton(
                onPressed: _busy ? null : _changePassword,
                child: const Text('Update password'),
              ),
            ],
          ),
        ),
        const SizedBox(height: AppConstants.sectionSpacing),
        _sectionCard(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text(
                'Danger Zone',
                style: TextStyle(
                  fontWeight: FontWeight.w700,
                  color: Colors.red,
                ),
              ),
              const SizedBox(height: 8),
              TextField(
                controller: _disablePasswordController,
                obscureText: true,
                decoration: const InputDecoration(
                  labelText: 'Password to disable account',
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 12),
              FilledButton(
                onPressed: _busy ? null : _disableAccount,
                style: FilledButton.styleFrom(backgroundColor: Colors.red),
                child: const Text('Disable account'),
              ),
            ],
          ),
        ),
      ],
    );
  }
}
