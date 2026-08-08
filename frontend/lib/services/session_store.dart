import 'dart:convert';

import 'package:shared_preferences/shared_preferences.dart';

import '../models.dart';

class SessionStore {
  static const String _sessionKey = 'odessa.auth.session';

  Future<void> save(AuthSession session) async {
    final prefs = await SharedPreferences.getInstance();
    final json = jsonEncode({
      'access_token': session.accessToken,
      'refresh_token': session.refreshToken,
      'token_type': session.tokenType,
      'expires_in': session.expiresIn,
    });
    await prefs.setString(_sessionKey, json);
  }

  Future<AuthSession?> load() async {
    final prefs = await SharedPreferences.getInstance();
    final raw = prefs.getString(_sessionKey);
    if (raw == null || raw.isEmpty) {
      return null;
    }

    try {
      final decoded = jsonDecode(raw);
      if (decoded is! Map<String, dynamic>) {
        return null;
      }

      final accessToken = (decoded['access_token'] as String?) ?? '';
      final refreshToken = (decoded['refresh_token'] as String?) ?? '';
      final tokenType = (decoded['token_type'] as String?) ?? 'Bearer';
      final expiresIn = (decoded['expires_in'] as num?)?.toInt() ?? 0;

      if (accessToken.isEmpty || refreshToken.isEmpty) {
        return null;
      }

      return AuthSession(
        accessToken: accessToken,
        refreshToken: refreshToken,
        tokenType: tokenType,
        expiresIn: expiresIn,
      );
    } catch (_) {
      return null;
    }
  }

  Future<void> clear() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_sessionKey);
  }
}
