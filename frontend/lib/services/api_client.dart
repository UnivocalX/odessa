import 'dart:convert';

import 'package:http/http.dart' as http;

import '../models.dart';

class ApiClient {
  ApiClient({String? baseUrl})
    : baseUrl =
          (baseUrl ??
                  const String.fromEnvironment(
                    'ODESSA_API_BASE_URL',
                    defaultValue: 'http://localhost:9090',
                  ))
              .replaceFirst(RegExp(r'/+$'), '');

  final String baseUrl;

  Future<AuthSession> login({
    required String email,
    required String password,
  }) async {
    final payload = await _requestJson(
      method: 'POST',
      path: '/api/auth/login',
      body: {'email': email, 'password': password},
      fallbackError: 'Login failed',
    );

    return AuthSession(
      accessToken: (payload['access_token'] as String?) ?? '',
      refreshToken: (payload['refresh_token'] as String?) ?? '',
      tokenType: (payload['token_type'] as String?) ?? 'Bearer',
      expiresIn: (payload['expires_in'] as num?)?.toInt() ?? 0,
    );
  }

  Future<AuthSession> refreshSession(String refreshToken) async {
    final payload = await _requestJson(
      method: 'POST',
      path: '/api/auth/refresh',
      body: {'refresh_token': refreshToken},
      fallbackError: 'Session refresh failed',
    );

    return AuthSession(
      accessToken: (payload['access_token'] as String?) ?? '',
      refreshToken: (payload['refresh_token'] as String?) ?? '',
      tokenType: (payload['token_type'] as String?) ?? 'Bearer',
      expiresIn: (payload['expires_in'] as num?)?.toInt() ?? 0,
    );
  }

  Future<String> logout({
    required String accessToken,
    required String refreshToken,
  }) async {
    final payload = await _requestJson(
      method: 'POST',
      path: '/api/auth/logout',
      accessToken: accessToken,
      body: {'refresh_token': refreshToken},
      fallbackError: 'Logout failed',
    );

    return (payload['message'] as String?) ?? 'logged out';
  }

  Future<String> changePassword({
    required String accessToken,
    required String currentPassword,
    required String newPassword,
  }) async {
    final payload = await _requestJson(
      method: 'POST',
      path: '/api/auth/password/change',
      accessToken: accessToken,
      body: {'current_password': currentPassword, 'new_password': newPassword},
      fallbackError: 'Change password failed',
    );

    return (payload['message'] as String?) ?? 'password changed';
  }

  Future<String> disableAccount({
    required String accessToken,
    required String password,
  }) async {
    final payload = await _requestJson(
      method: 'POST',
      path: '/api/auth/account/disable',
      accessToken: accessToken,
      body: {'password': password},
      fallbackError: 'Disable account failed',
    );

    return (payload['message'] as String?) ?? 'account disabled';
  }

  Future<String> requestPasswordReset(String email) async {
    final payload = await _requestJson(
      method: 'POST',
      path: '/api/auth/password/reset/request',
      body: {'email': email},
      fallbackError: 'Password reset request failed',
    );

    return (payload['message'] as String?) ?? 'password reset requested';
  }

  Future<String> confirmPasswordReset({
    required String token,
    required String password,
  }) async {
    final payload = await _requestJson(
      method: 'POST',
      path: '/api/auth/password/reset/confirm',
      body: {'token': token, 'password': password},
      fallbackError: 'Password reset confirm failed',
    );

    return (payload['message'] as String?) ?? 'password reset confirmed';
  }

  Future<List<DatasetItem>> getDatasets(String accessToken) async {
    final payload = await _requestJson(
      method: 'GET',
      path: '/api/v1/datasets',
      accessToken: accessToken,
      fallbackError: 'Failed to load datasets',
    );

    final datasets = (payload['datasets'] as List<dynamic>? ?? const [])
        .map((item) => _toMap(item))
        .whereType<Map<String, dynamic>>()
        .map((item) => DatasetItem.fromJson(item))
        .toList();

    return datasets;
  }

  Future<DatasetItem> createDataset({
    required String accessToken,
    required String name,
    required String description,
  }) async {
    final payload = await _requestJson(
      method: 'POST',
      path: '/api/v1/datasets',
      accessToken: accessToken,
      body: {'name': name, 'description': description},
      fallbackError: 'Failed to create dataset',
    );

    return DatasetItem.fromJson(payload);
  }

  Future<List<DatasetVersionItem>> getDatasetVersions({
    required String accessToken,
    required int datasetId,
  }) async {
    final payload = await _requestJson(
      method: 'GET',
      path: '/api/v1/datasets/$datasetId/versions',
      accessToken: accessToken,
      fallbackError: 'Failed to load dataset versions',
    );

    final versions = (payload['versions'] as List<dynamic>? ?? const [])
        .map((item) => _toMap(item))
        .whereType<Map<String, dynamic>>()
        .map((item) => DatasetVersionItem.fromJson(item))
        .toList();

    return versions;
  }

  Future<DatasetVersionItem> createDatasetVersion({
    required String accessToken,
    required int datasetId,
    required String commit,
    required List<int> blobIds,
  }) async {
    final payload = await _requestJson(
      method: 'POST',
      path: '/api/v1/datasets/$datasetId/versions',
      accessToken: accessToken,
      body: {'commit': commit, 'blob_ids': blobIds},
      fallbackError: 'Failed to create dataset version',
    );

    return DatasetVersionItem.fromJson(payload);
  }

  Future<DatasetVersionItem> getDatasetVersion({
    required String accessToken,
    required int datasetId,
    required int versionId,
  }) async {
    final payload = await _requestJson(
      method: 'GET',
      path: '/api/v1/datasets/$datasetId/versions/$versionId',
      accessToken: accessToken,
      fallbackError: 'Failed to load dataset version',
    );

    return DatasetVersionItem.fromJson(payload);
  }

  Future<BlobSearchPageResult> getDatasetVersionBlobs({
    required String accessToken,
    required int datasetId,
    required int versionId,
    int cursor = 0,
    int limit = 100,
  }) async {
    final payload = await _requestJson(
      method: 'GET',
      path:
          '/api/v1/datasets/$datasetId/versions/$versionId/blobs?cursor=$cursor&limit=$limit',
      accessToken: accessToken,
      fallbackError: 'Failed to load dataset version blobs',
    );

    final blobs = (payload['blobs'] as List<dynamic>? ?? const [])
        .map((item) => _toMap(item))
        .whereType<Map<String, dynamic>>()
        .map((item) => BlobItem.fromJson(item))
        .toList();

    return BlobSearchPageResult(
      blobs: blobs,
      hasMore: (payload['has_more'] as bool?) ?? false,
      nextCursor: (payload['next_cursor'] as num?)?.toInt() ?? 0,
      total: (payload['total'] as num?)?.toInt() ?? blobs.length,
    );
  }

  Future<List<BlobItem>> getBlobs(String accessToken) async {
    final payload = await _requestJson(
      method: 'GET',
      path: '/api/v1/blobs?limit=100',
      accessToken: accessToken,
      fallbackError: 'Failed to load blobs',
    );

    final blobs = (payload['blobs'] as List<dynamic>? ?? const [])
        .map((item) => _toMap(item))
        .whereType<Map<String, dynamic>>()
        .map((item) => BlobItem.fromJson(item))
        .toList();

    return blobs;
  }

  Future<BlobItem> getBlobByHash({
    required String accessToken,
    required String hash,
  }) async {
    final payload = await _requestJson(
      method: 'GET',
      path: '/api/v1/blobs/$hash',
      accessToken: accessToken,
      fallbackError: 'Failed to load blob',
    );

    return BlobItem.fromJson(payload);
  }

  Future<BlobItem> getBlobById({
    required String accessToken,
    required int id,
  }) async {
    final payload = await _requestJson(
      method: 'GET',
      path: '/api/v1/blobs/id/$id',
      accessToken: accessToken,
      fallbackError: 'Failed to load blob',
    );

    return BlobItem.fromJson(payload);
  }

  Future<BlobSearchPageResult> searchBlobs({
    required String accessToken,
    required Map<String, dynamic> query,
  }) async {
    final payload = await _requestJson(
      method: 'POST',
      path: '/api/v1/blobs/search',
      accessToken: accessToken,
      body: query,
      fallbackError: 'Failed to search blobs',
    );

    final blobs = (payload['blobs'] as List<dynamic>? ?? const [])
        .map((item) => _toMap(item))
        .whereType<Map<String, dynamic>>()
        .map((item) => BlobItem.fromJson(item))
        .toList();

    return BlobSearchPageResult(
      blobs: blobs,
      hasMore: (payload['has_more'] as bool?) ?? false,
      nextCursor: (payload['next_cursor'] as num?)?.toInt() ?? 0,
      total: (payload['total'] as num?)?.toInt() ?? blobs.length,
    );
  }

  Future<List<OriginItem>> getOrigins(String accessToken) async {
    final payload = await _requestJson(
      method: 'GET',
      path: '/api/v1/origins',
      accessToken: accessToken,
      fallbackError: 'Failed to load origins',
    );

    final origins = (payload['origins'] as List<dynamic>? ?? const [])
        .map((item) => _toMap(item))
        .whereType<Map<String, dynamic>>()
        .map((item) => OriginItem.fromJson(item))
        .toList();

    return origins;
  }

  Future<String> createOrigin({
    required String accessToken,
    required String uri,
    Map<String, dynamic>? rules,
  }) async {
    final payload = await _requestJson(
      method: 'POST',
      path: '/api/v1/origins',
      accessToken: accessToken,
      body: {'uri': uri, if (rules != null) 'rules': rules},
      fallbackError: 'Failed to create origin',
    );

    return (payload['message'] as String?) ?? 'origin created';
  }

  Future<String> updateOriginRules({
    required String accessToken,
    required int originId,
    required Map<String, dynamic> rules,
  }) async {
    final payload = await _requestJson(
      method: 'PUT',
      path: '/api/v1/origins/$originId/rules',
      accessToken: accessToken,
      body: {'rules': rules},
      fallbackError: 'Failed to update rules',
    );

    return (payload['message'] as String?) ?? 'rules updated';
  }

  Future<ScanOriginItem> triggerOriginScan({
    required String accessToken,
    required int originId,
    Map<String, dynamic>? rules,
  }) async {
    final payload = await _requestJson(
      method: 'POST',
      path: '/api/v1/origins/$originId/scan',
      accessToken: accessToken,
      body: {if (rules != null) 'rules': rules},
      fallbackError: 'Failed to trigger origin scan',
    );

    return ScanOriginItem.fromJson(payload);
  }

  Future<ScanOriginItem> getOriginScan({
    required String accessToken,
    required int originId,
  }) async {
    final payload = await _requestJson(
      method: 'GET',
      path: '/api/v1/origins/$originId/scan',
      accessToken: accessToken,
      fallbackError: 'Failed to get origin scan',
    );

    return ScanOriginItem.fromJson(payload);
  }

  Future<ScanOriginItem> cancelOriginScan({
    required String accessToken,
    required int originId,
  }) async {
    final payload = await _requestJson(
      method: 'POST',
      path: '/api/v1/origins/$originId/scan/cancel',
      accessToken: accessToken,
      fallbackError: 'Failed to cancel origin scan',
    );

    return ScanOriginItem.fromJson(payload);
  }

  Future<List<LabelItem>> getLabels(String accessToken) async {
    final payload = await _requestJson(
      method: 'GET',
      path: '/api/v1/labels',
      accessToken: accessToken,
      fallbackError: 'Failed to load labels',
    );

    final labels = (payload['labels'] as List<dynamic>? ?? const [])
        .map((item) => _toMap(item))
        .whereType<Map<String, dynamic>>()
        .map((item) => LabelItem.fromJson(item))
        .toList();

    return labels;
  }

  Future<LabelItem> createLabel({
    required String accessToken,
    required String name,
    required String description,
  }) async {
    final payload = await _requestJson(
      method: 'POST',
      path: '/api/v1/labels',
      accessToken: accessToken,
      body: {'name': name, 'description': description},
      fallbackError: 'Failed to create label',
    );

    return LabelItem.fromJson(payload);
  }

  Future<List<UserItem>> getUsers(String accessToken) async {
    final payload = await _requestJson(
      method: 'GET',
      path: '/api/v1/users',
      accessToken: accessToken,
      fallbackError: 'Failed to load users',
    );

    final users = (payload['users'] as List<dynamic>? ?? const [])
        .map((item) => _toMap(item))
        .whereType<Map<String, dynamic>>()
        .map((item) => UserItem.fromJson(item))
        .toList();

    return users;
  }

  Future<String> createUser({
    required String accessToken,
    required String name,
    required String email,
    required String password,
  }) async {
    final payload = await _requestJson(
      method: 'POST',
      path: '/api/v1/users',
      accessToken: accessToken,
      body: {'name': name, 'email': email, 'password': password},
      fallbackError: 'Failed to create user',
    );

    return (payload['message'] as String?) ?? 'user created';
  }

  Future<void> deleteUser({
    required String accessToken,
    required int userId,
  }) async {
    await _requestJson(
      method: 'DELETE',
      path: '/api/v1/users/$userId',
      accessToken: accessToken,
      fallbackError: 'Failed to delete user',
      expectNoContent: true,
    );
  }

  Map<String, String> _authHeaders(String accessToken) {
    return {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer $accessToken',
    };
  }

  Future<Map<String, dynamic>> _requestJson({
    required String method,
    required String path,
    String? accessToken,
    Map<String, dynamic>? body,
    required String fallbackError,
    bool expectNoContent = false,
  }) async {
    final uri = Uri.parse('$baseUrl$path');
    final headers = accessToken == null
        ? const {'Content-Type': 'application/json'}
        : _authHeaders(accessToken);

    late final http.Response response;
    switch (method) {
      case 'GET':
        response = await http.get(uri, headers: headers);
        break;
      case 'POST':
        response = await http.post(
          uri,
          headers: headers,
          body: body == null ? null : jsonEncode(body),
        );
        break;
      case 'PUT':
        response = await http.put(
          uri,
          headers: headers,
          body: body == null ? null : jsonEncode(body),
        );
        break;
      case 'DELETE':
        response = await http.delete(uri, headers: headers);
        break;
      default:
        throw ApiException('Unsupported method: $method');
    }

    if (expectNoContent && response.statusCode == 204) {
      return <String, dynamic>{};
    }

    final payload = _decodeJson(response.body);
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw ApiException(_extractError(payload, fallback: fallbackError));
    }

    return payload;
  }

  Map<String, dynamic>? _toMap(dynamic input) {
    if (input is Map<String, dynamic>) {
      return input;
    }
    if (input is Map) {
      return input.map((key, value) => MapEntry('$key', value));
    }
    return null;
  }

  Map<String, dynamic> _decodeJson(String body) {
    if (body.isEmpty) {
      return <String, dynamic>{};
    }

    final decoded = jsonDecode(body);
    if (decoded is Map<String, dynamic>) {
      return decoded;
    }

    return <String, dynamic>{};
  }

  String _extractError(
    Map<String, dynamic> payload, {
    required String fallback,
  }) {
    final error = payload['error'];
    final message = payload['message'];
    if (error is String && error.isNotEmpty) {
      return error;
    }
    if (message is String && message.isNotEmpty) {
      return message;
    }
    return fallback;
  }
}

class ApiException implements Exception {
  ApiException(this.message);

  final String message;

  @override
  String toString() => message;
}
