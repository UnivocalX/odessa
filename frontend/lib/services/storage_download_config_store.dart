import 'dart:convert';

import 'package:shared_preferences/shared_preferences.dart';

class StorageDownloadConfig {
  const StorageDownloadConfig({
    required this.downloadUrlTemplate,
    required this.s3ApiEndpoint,
    required this.bearerToken,
    required this.accessKey,
    required this.secretKey,
    required this.customHeaderName,
    required this.customHeaderValue,
  });

  static const StorageDownloadConfig empty = StorageDownloadConfig(
    downloadUrlTemplate: '',
    s3ApiEndpoint: '',
    bearerToken: '',
    accessKey: '',
    secretKey: '',
    customHeaderName: '',
    customHeaderValue: '',
  );

  final String downloadUrlTemplate;
  final String s3ApiEndpoint;
  final String bearerToken;
  final String accessKey;
  final String secretKey;
  final String customHeaderName;
  final String customHeaderValue;

  StorageDownloadConfig copyWith({
    String? downloadUrlTemplate,
    String? s3ApiEndpoint,
    String? bearerToken,
    String? accessKey,
    String? secretKey,
    String? customHeaderName,
    String? customHeaderValue,
  }) {
    return StorageDownloadConfig(
      downloadUrlTemplate: downloadUrlTemplate ?? this.downloadUrlTemplate,
      s3ApiEndpoint: s3ApiEndpoint ?? this.s3ApiEndpoint,
      bearerToken: bearerToken ?? this.bearerToken,
      accessKey: accessKey ?? this.accessKey,
      secretKey: secretKey ?? this.secretKey,
      customHeaderName: customHeaderName ?? this.customHeaderName,
      customHeaderValue: customHeaderValue ?? this.customHeaderValue,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'download_url_template': downloadUrlTemplate,
      's3_api_endpoint': s3ApiEndpoint,
      'bearer_token': bearerToken,
      'access_key': accessKey,
      'secret_key': secretKey,
      'custom_header_name': customHeaderName,
      'custom_header_value': customHeaderValue,
    };
  }

  factory StorageDownloadConfig.fromJson(Map<String, dynamic> json) {
    return StorageDownloadConfig(
      downloadUrlTemplate: (json['download_url_template'] as String?) ?? '',
      s3ApiEndpoint: (json['s3_api_endpoint'] as String?) ?? '',
      bearerToken: (json['bearer_token'] as String?) ?? '',
      accessKey: (json['access_key'] as String?) ?? '',
      secretKey: (json['secret_key'] as String?) ?? '',
      customHeaderName: (json['custom_header_name'] as String?) ?? '',
      customHeaderValue: (json['custom_header_value'] as String?) ?? '',
    );
  }
}

class StorageDownloadConfigStore {
  static const String _storageKey = 'odessa_storage_download_config';

  Future<StorageDownloadConfig> load() async {
    final preferences = await SharedPreferences.getInstance();
    final raw = preferences.getString(_storageKey);
    if (raw == null || raw.trim().isEmpty) {
      return StorageDownloadConfig.empty;
    }

    try {
      final decoded = jsonDecode(raw);
      if (decoded is! Map<String, dynamic>) {
        return StorageDownloadConfig.empty;
      }
      return StorageDownloadConfig.fromJson(decoded);
    } catch (_) {
      return StorageDownloadConfig.empty;
    }
  }

  Future<void> save(StorageDownloadConfig config) async {
    final preferences = await SharedPreferences.getInstance();
    await preferences.setString(_storageKey, jsonEncode(config.toJson()));
  }
}
