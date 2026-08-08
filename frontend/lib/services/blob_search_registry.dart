import 'dart:convert';

import 'package:flutter/foundation.dart';
import 'package:shared_preferences/shared_preferences.dart';

class BlobSearchSnapshot {
  const BlobSearchSnapshot({
    required this.id,
    required this.name,
    required this.createdAt,
    required this.blobIds,
    required this.resultCount,
    required this.filters,
  });

  final String id;
  final String name;
  final DateTime createdAt;
  final List<int> blobIds;
  final int resultCount;
  final Map<String, dynamic> filters;

  BlobSearchSnapshot copyWith({
    String? id,
    String? name,
    DateTime? createdAt,
    List<int>? blobIds,
    int? resultCount,
    Map<String, dynamic>? filters,
  }) {
    return BlobSearchSnapshot(
      id: id ?? this.id,
      name: name ?? this.name,
      createdAt: createdAt ?? this.createdAt,
      blobIds: blobIds ?? this.blobIds,
      resultCount: resultCount ?? this.resultCount,
      filters: filters ?? this.filters,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'name': name,
      'created_at': createdAt.toIso8601String(),
      'blob_ids': blobIds,
      'result_count': resultCount,
      'filters': filters,
    };
  }

  factory BlobSearchSnapshot.fromJson(Map<String, dynamic> json) {
    final createdAtRaw = (json['created_at'] as String?) ?? '';
    final createdAt = DateTime.tryParse(createdAtRaw) ?? DateTime.now();
    return BlobSearchSnapshot(
      id: (json['id'] as String?) ?? '',
      name: (json['name'] as String?) ?? '',
      createdAt: createdAt,
      blobIds: (json['blob_ids'] as List<dynamic>? ?? const [])
          .whereType<num>()
          .map((item) => item.toInt())
          .toList(),
      resultCount: (json['result_count'] as num?)?.toInt() ?? 0,
      filters: Map<String, dynamic>.from(
        (json['filters'] as Map<String, dynamic>?) ?? const {},
      ),
    );
  }
}

class BlobSearchRegistry extends ChangeNotifier {
  BlobSearchRegistry._();

  static final BlobSearchRegistry instance = BlobSearchRegistry._();

  static const String _storageKey = 'odessa_blob_search_snapshots';

  final List<BlobSearchSnapshot> _snapshots = [];
  bool _loaded = false;

  List<BlobSearchSnapshot> get snapshots => List.unmodifiable(_snapshots);

  bool get isLoaded => _loaded;

  Future<void> load() async {
    if (_loaded) {
      return;
    }
    final preferences = await SharedPreferences.getInstance();
    final raw = preferences.getString(_storageKey);
    if (raw == null || raw.trim().isEmpty) {
      _loaded = true;
      notifyListeners();
      return;
    }

    try {
      final decoded = jsonDecode(raw);
      if (decoded is List) {
        _snapshots
          ..clear()
          ..addAll(
            decoded.whereType<Map>().map(
              (item) => BlobSearchSnapshot.fromJson(
                item.map((key, value) => MapEntry('$key', value)),
              ),
            ),
          );
      }
    } catch (_) {
      _snapshots.clear();
    }

    _loaded = true;
    notifyListeners();
  }

  Future<void> _persist() async {
    final preferences = await SharedPreferences.getInstance();
    final payload = jsonEncode(
      _snapshots.map((item) => item.toJson()).toList(),
    );
    await preferences.setString(_storageKey, payload);
  }

  Future<void> add(BlobSearchSnapshot snapshot) async {
    await load();
    _snapshots.insert(0, snapshot);
    if (_snapshots.length > 100) {
      _snapshots.removeRange(100, _snapshots.length);
    }
    await _persist();
    notifyListeners();
  }

  Future<void> upsert(BlobSearchSnapshot snapshot) async {
    await load();
    final existing = _snapshots.indexWhere((item) => item.id == snapshot.id);
    if (existing >= 0) {
      _snapshots[existing] = snapshot;
      if (existing > 0) {
        final updated = _snapshots.removeAt(existing);
        _snapshots.insert(0, updated);
      }
      await _persist();
      notifyListeners();
      return;
    }

    await add(snapshot);
  }

  Future<void> rename(String id, String name) async {
    await load();
    final trimmed = name.trim();
    if (trimmed.isEmpty) {
      return;
    }

    final existing = _snapshots.indexWhere((item) => item.id == id);
    if (existing < 0) {
      return;
    }

    _snapshots[existing] = _snapshots[existing].copyWith(name: trimmed);
    await _persist();
    notifyListeners();
  }

  Future<void> remove(String id) async {
    await load();
    final before = _snapshots.length;
    _snapshots.removeWhere((item) => item.id == id);
    if (_snapshots.length == before) {
      return;
    }

    await _persist();
    notifyListeners();
  }
}
