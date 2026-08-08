class AuthSession {
  const AuthSession({
    required this.accessToken,
    required this.refreshToken,
    required this.tokenType,
    required this.expiresIn,
  });

  final String accessToken;
  final String refreshToken;
  final String tokenType;
  final int expiresIn;
}

class DatasetItem {
  const DatasetItem({
    required this.id,
    required this.name,
    required this.description,
    required this.createdAt,
  });

  final int id;
  final String name;
  final String description;
  final String createdAt;

  factory DatasetItem.fromJson(Map<String, dynamic> json) {
    return DatasetItem(
      id: (json['id'] as num?)?.toInt() ?? 0,
      name: (json['name'] as String?) ?? '',
      description: (json['description'] as String?) ?? '',
      createdAt: (json['created_at'] as String?) ?? '',
    );
  }
}

class BlobItem {
  const BlobItem({
    required this.id,
    required this.hash,
    required this.mimeType,
    required this.size,
    required this.locations,
    required this.labels,
  });

  final int id;
  final String hash;
  final String mimeType;
  final int size;
  final List<String> locations;
  final List<BlobLabelItem> labels;

  factory BlobItem.fromJson(Map<String, dynamic> json) {
    return BlobItem(
      id: (json['id'] as num?)?.toInt() ?? 0,
      hash: (json['hash'] as String?) ?? '',
      mimeType: (json['mime_type'] as String?) ?? '',
      size: (json['size'] as num?)?.toInt() ?? 0,
      locations: (json['locations'] as List<dynamic>? ?? const [])
          .map((value) => '$value')
          .where((value) => value.isNotEmpty)
          .toList(),
      labels: (json['labels'] as List<dynamic>? ?? const [])
          .whereType<Map>()
          .map((value) => Map<String, dynamic>.from(value))
          .whereType<Map<String, dynamic>>()
          .map((value) => BlobLabelItem.fromJson(value))
          .toList(),
    );
  }

  Map<String, dynamic> toJson() {
    return <String, dynamic>{
      'id': id,
      'hash': hash,
      'mime_type': mimeType,
      'size': size,
      'locations': locations,
      'labels': labels.map((item) => item.toJson()).toList(),
    };
  }
}

class BlobLabelItem {
  const BlobLabelItem({required this.name, required this.value});

  final String name;
  final String value;

  factory BlobLabelItem.fromJson(Map<String, dynamic> json) {
    return BlobLabelItem(
      name: (json['name'] as String?) ?? '',
      value: (json['value'] as String?) ?? '',
    );
  }

  Map<String, dynamic> toJson() {
    return <String, dynamic>{'name': name, 'value': value};
  }
}

class OriginItem {
  const OriginItem({
    required this.id,
    required this.uri,
    required this.createdAt,
  });

  final int id;
  final String uri;
  final String createdAt;

  factory OriginItem.fromJson(Map<String, dynamic> json) {
    return OriginItem(
      id: (json['id'] as num?)?.toInt() ?? 0,
      uri: (json['uri'] as String?) ?? '',
      createdAt: (json['created_at'] as String?) ?? '',
    );
  }
}

class LabelItem {
  const LabelItem({
    required this.id,
    required this.name,
    required this.description,
  });

  final int id;
  final String name;
  final String description;

  factory LabelItem.fromJson(Map<String, dynamic> json) {
    return LabelItem(
      id: (json['id'] as num?)?.toInt() ?? 0,
      name: ((json['name'] ?? json['Name']) as String?) ?? '',
      description: (json['description'] as String?) ?? '',
    );
  }
}

class UserItem {
  const UserItem({
    required this.id,
    required this.name,
    required this.email,
    required this.role,
    required this.createdAt,
    required this.disabledAt,
  });

  final int id;
  final String name;
  final String email;
  final String role;
  final String createdAt;
  final String? disabledAt;

  factory UserItem.fromJson(Map<String, dynamic> json) {
    return UserItem(
      id: (json['id'] as num?)?.toInt() ?? 0,
      name: (json['name'] as String?) ?? '',
      email: (json['email'] as String?) ?? '',
      role: (json['role'] as String?) ?? '',
      createdAt: (json['created_at'] as String?) ?? '',
      disabledAt: json['disabled_at'] as String?,
    );
  }
}

class DatasetVersionItem {
  const DatasetVersionItem({
    required this.id,
    required this.datasetId,
    required this.commit,
    required this.createdAt,
    required this.blobIds,
  });

  final int id;
  final int datasetId;
  final String commit;
  final String createdAt;
  final List<int> blobIds;

  factory DatasetVersionItem.fromJson(Map<String, dynamic> json) {
    return DatasetVersionItem(
      id: (json['id'] as num?)?.toInt() ?? 0,
      datasetId: (json['dataset_id'] as num?)?.toInt() ?? 0,
      commit: (json['commit'] as String?) ?? '',
      createdAt: (json['created_at'] as String?) ?? '',
      blobIds: (json['blob_ids'] as List<dynamic>? ?? const [])
          .whereType<num>()
          .map((value) => value.toInt())
          .toList(),
    );
  }
}

class BlobSearchPageResult {
  const BlobSearchPageResult({
    required this.blobs,
    required this.hasMore,
    required this.nextCursor,
    required this.total,
  });

  final List<BlobItem> blobs;
  final bool hasMore;
  final int nextCursor;
  final int total;

  int get pageCount => blobs.length;
}

class ScanOriginItem {
  const ScanOriginItem({
    required this.id,
    required this.originId,
    required this.status,
    required this.attempts,
    required this.results,
    required this.createdAt,
  });

  final int id;
  final int originId;
  final String status;
  final int attempts;
  final Map<String, dynamic> results;
  final String createdAt;

  int get discoveredFiles {
    final raw = results['discovered'];
    if (raw is num) {
      final value = raw.toInt();
      return value < 0 ? 0 : value;
    }
    return 0;
  }

  factory ScanOriginItem.fromJson(Map<String, dynamic> json) {
    final parsedResults =
        (json['results'] as Map<String, dynamic>?) ?? const {};
    return ScanOriginItem(
      id: (json['id'] as num?)?.toInt() ?? 0,
      originId: (json['origin_id'] as num?)?.toInt() ?? 0,
      status: (json['status'] as String?) ?? '',
      attempts: (json['attempts'] as num?)?.toInt() ?? 0,
      results: Map<String, dynamic>.from(parsedResults),
      createdAt: (json['created_at'] as String?) ?? '',
    );
  }
}
