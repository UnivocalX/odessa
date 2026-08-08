import 'dart:async';

import 'package:flutter/material.dart';

import '../../core/app_constants.dart';
import '../../models.dart';
import '../../services/api_client.dart';
import '../../services/blob_search_registry.dart';

class BlobsView extends StatefulWidget {
  const BlobsView({super.key, required this.apiClient, required this.session});
  final ApiClient apiClient;
  final AuthSession session;

  @override
  State<BlobsView> createState() => _BlobsViewState();
}

class _BlobsViewState extends State<BlobsView> {
  static const int _pageSize = AppConstants.blobSearchPageSize;
  static const List<String> _fileTypeOptions = <String>[
    'image/png',
    'image/jpeg',
    'image/gif',
    'image/tiff',
    'image/svg+xml',
    'image/webp',
    'application/pdf',
    'text/plain',
    'text/csv',
    'text/markdown',
    'text/html',
    'text/xml',
    'application/json',
    'application/xml',
    'application/zip',
    'application/gzip',
    'application/x-tar',
    'application/msword',
    'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
    'application/vnd.ms-excel',
    'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    'application/vnd.ms-powerpoint',
    'application/vnd.openxmlformats-officedocument.presentationml.presentation',
    'audio/mpeg',
    'audio/wav',
    'audio/ogg',
    'video/mp4',
    'video/quicktime',
    'video/x-msvideo',
  ];

  static const List<_SizePreset> _sizePresets = <_SizePreset>[
    _SizePreset(id: 'tiny', label: 'Tiny (0-100 KB)', min: 0, max: 100000),
    _SizePreset(
      id: 'small',
      label: 'Small (100 KB-1 MB)',
      min: 100000,
      max: 1000000,
    ),
    _SizePreset(
      id: 'medium',
      label: 'Medium (1 MB-10 MB)',
      min: 1000000,
      max: 10000000,
    ),
    _SizePreset(id: 'large', label: 'Large (10 MB+)', min: 10000000, max: null),
  ];

  bool _searching = false;
  bool _hasSearched = false;
  int _currentPage = 1;
  final TextEditingController _searchNameController = TextEditingController();
  final List<_FilterRule> _activeFilters = [];
  Map<String, dynamic>? _latestQuery;
  List<BlobItem>? _latestResults;
  String? _activeSnapshotId;
  DateTime? _activeSnapshotCreatedAt;
  Map<String, dynamic>? _activeSnapshotBaseQuery;
  bool _hasMoreResults = false;
  int _nextCursor = 0;
  int _latestTotalMatches = 0;
  final List<int> _pageCursors = [0];
  int _searchRequestSerial = 0;

  @override
  void initState() {
    super.initState();
    _searchNameController.text = _defaultSearchName();
    BlobSearchRegistry.instance.addListener(_onSnapshotsChanged);
  }

  void _onSnapshotsChanged() {
    if (mounted) {
      setState(() {});
    }
  }

  @override
  void dispose() {
    BlobSearchRegistry.instance.removeListener(_onSnapshotsChanged);
    _searchNameController.dispose();
    super.dispose();
  }

  String _defaultSearchName() {
    final now = DateTime.now();
    final month = now.month.toString().padLeft(2, '0');
    final day = now.day.toString().padLeft(2, '0');
    final hour = now.hour.toString().padLeft(2, '0');
    final minute = now.minute.toString().padLeft(2, '0');
    return 'Search ${now.year}-$month-$day $hour:$minute';
  }

  List<String> _parseCsv(String text) {
    return text
        .split(',')
        .map((value) => value.trim())
        .where((value) => value.isNotEmpty)
        .toList();
  }

  _LabelFilterSplit _parseLabelFilter(String text) {
    final labels = <String>[];
    final labelValues = <String, String>{};

    for (final token in text.split(',')) {
      final item = token.trim();
      if (item.isEmpty) {
        continue;
      }
      final idx = item.indexOf('=');
      if (idx > 0 && idx < item.length - 1) {
        final key = item.substring(0, idx).trim();
        final value = item.substring(idx + 1).trim();
        if (key.isNotEmpty && value.isNotEmpty) {
          labelValues[key] = value;
        }
      } else {
        labels.add(item);
      }
    }

    return _LabelFilterSplit(labels: labels, labelValues: labelValues);
  }

  void _appendUniqueList(
    Map<String, dynamic> target,
    String key,
    List<String> values,
  ) {
    if (values.isEmpty) {
      return;
    }

    final existing = (target[key] as List<dynamic>? ?? const [])
        .whereType<String>()
        .toSet();
    existing.addAll(values);
    target[key] = existing.toList();
  }

  void _mergeLabelValues(
    Map<String, dynamic> target,
    Map<String, String> values,
  ) {
    if (values.isEmpty) {
      return;
    }

    final merged = <String, String>{
      ...(target['label_values'] as Map<String, dynamic>? ?? const {}).map(
        (key, value) => MapEntry('$key', '$value'),
      ),
      ...values,
    };
    target['label_values'] = merged;
  }

  Map<String, dynamic> _buildQuery() {
    final includeFilter = <String, dynamic>{};
    final excludeFilter = <String, dynamic>{};
    int? minSize;
    int? maxSize;

    for (final filter in _activeFilters) {
      if (filter.kind == _FilterKind.sizeRange) {
        final preset = _sizePresets
            .where((item) => item.id == filter.value)
            .cast<_SizePreset?>()
            .firstWhere((item) => item != null, orElse: () => null);
        if (preset != null) {
          minSize = preset.min;
          maxSize = preset.max;
        }
        continue;
      }

      final target = filter.include ? includeFilter : excludeFilter;
      switch (filter.kind) {
        case _FilterKind.checksums:
          final values = _parseCsv(filter.value);
          _appendUniqueList(target, 'hashes', values);
          break;
        case _FilterKind.fileTypes:
          final values = _parseCsv(filter.value);
          _appendUniqueList(target, 'mime_types', values);
          break;
        case _FilterKind.labels:
          final split = _parseLabelFilter(filter.value);
          _appendUniqueList(target, 'labels', split.labels);
          _mergeLabelValues(target, split.labelValues);
          break;
        case _FilterKind.pathPattern:
          final pattern = filter.value.trim();
          if (pattern.isNotEmpty) {
            target['uri_pattern'] = pattern;
          }
          break;
        case _FilterKind.sizeRange:
          break;
      }
    }

    final query = <String, dynamic>{};
    if (includeFilter.isNotEmpty) {
      query['include'] = includeFilter;
    }
    if (excludeFilter.isNotEmpty) {
      query['exclude'] = excludeFilter;
    }
    if (minSize != null) {
      query['min_size'] = minSize;
    }
    if (maxSize != null) {
      query['max_size'] = maxSize;
    }
    query['cursor'] = 0;
    query['limit'] = _pageSize;

    return query;
  }

  Map<String, dynamic> _queryWithoutPagination(Map<String, dynamic> query) {
    final base = Map<String, dynamic>.from(query);
    base.remove('cursor');
    base.remove('limit');
    return base;
  }

  Map<String, dynamic> _buildQueryWithPage(Map<String, dynamic> baseQuery) {
    final pageIndex = _currentPage - 1;
    final cursor = pageIndex >= 0 && pageIndex < _pageCursors.length
        ? _pageCursors[pageIndex]
        : 0;

    return <String, dynamic>{
      ...baseQuery,
      'cursor': cursor,
      'limit': _pageSize,
    };
  }

  int _pageCount() {
    if (_latestTotalMatches <= 0) {
      return 1;
    }
    return (_latestTotalMatches + _pageSize - 1) ~/ _pageSize;
  }

  String _formatBytes(int value) {
    if (value < 1024) {
      return '$value B';
    }
    if (value < 1024 * 1024) {
      return '${(value / 1024).toStringAsFixed(1)} KB';
    }
    if (value < 1024 * 1024 * 1024) {
      return '${(value / (1024 * 1024)).toStringAsFixed(1)} MB';
    }
    return '${(value / (1024 * 1024 * 1024)).toStringAsFixed(1)} GB';
  }

  Future<void> _resumeSavedSearch(BlobSearchSnapshot snapshot) async {
    final filters = Map<String, dynamic>.from(snapshot.filters);
    final cursor = (filters['cursor'] as num?)?.toInt() ?? 0;

    setState(() {
      _activeSnapshotId = snapshot.id;
      _activeSnapshotCreatedAt = snapshot.createdAt;
      _activeSnapshotBaseQuery = _queryWithoutPagination(filters);
      _searchNameController.text = snapshot.name;
      _currentPage = 1;
      _pageCursors
        ..clear()
        ..add(cursor < 0 ? 0 : cursor);
    });

    _showActionMessage('Loading saved search: ${snapshot.name}');
    await _search(preserveSnapshot: true);
  }

  Future<void> _showFiltersDialog() async {
    if (_latestQuery == null) {
      return;
    }

    final query = _latestQuery!;
    final include = Map<String, dynamic>.from(
      (query['include'] as Map<String, dynamic>?) ?? const {},
    );
    final exclude = Map<String, dynamic>.from(
      (query['exclude'] as Map<String, dynamic>?) ?? const {},
    );

    List<String> values(dynamic raw) {
      return (raw as List<dynamic>? ?? const [])
          .map((item) => '$item')
          .where((item) => item.isNotEmpty)
          .toList();
    }

    List<Widget> sectionRows(Map<String, dynamic> map) {
      final rows = <Widget>[];

      final hashes = values(map['hashes']);
      if (hashes.isNotEmpty) {
        rows.add(_filterRow('Hashes', hashes.join(', ')));
      }

      final mimeTypes = values(map['mime_types']);
      if (mimeTypes.isNotEmpty) {
        rows.add(_filterRow('File types', mimeTypes.join(', ')));
      }

      final labels = values(map['labels']);
      if (labels.isNotEmpty) {
        rows.add(_filterRow('Labels', labels.join(', ')));
      }

      final labelValues = Map<String, dynamic>.from(
        (map['label_values'] as Map<String, dynamic>?) ?? const {},
      );
      if (labelValues.isNotEmpty) {
        final formatted = labelValues.entries
            .map((entry) => '${entry.key}=${entry.value}')
            .join(', ');
        rows.add(_filterRow('Label values', formatted));
      }

      final pattern = '${map['uri_pattern'] ?? ''}'.trim();
      if (pattern.isNotEmpty) {
        rows.add(_filterRow('Path pattern', pattern));
      }

      return rows;
    }

    await showDialog<void>(
      context: context,
      builder: (context) {
        return AlertDialog(
          title: const Text('Search filters'),
          content: SizedBox(
            width: 520,
            child: SingleChildScrollView(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text(
                    'Include',
                    style: TextStyle(fontWeight: FontWeight.w700),
                  ),
                  const SizedBox(height: 6),
                  ...(sectionRows(include).isEmpty
                      ? [
                          Text(
                            'No include filters',
                            style: TextStyle(
                              color: Colors.white.withValues(alpha: 0.72),
                            ),
                          ),
                        ]
                      : sectionRows(include)),
                  const SizedBox(height: 12),
                  const Text(
                    'Exclude',
                    style: TextStyle(fontWeight: FontWeight.w700),
                  ),
                  const SizedBox(height: 6),
                  ...(sectionRows(exclude).isEmpty
                      ? [
                          Text(
                            'No exclude filters',
                            style: TextStyle(
                              color: Colors.white.withValues(alpha: 0.72),
                            ),
                          ),
                        ]
                      : sectionRows(exclude)),
                  const SizedBox(height: 12),
                  _filterRow('Min size', '${query['min_size'] ?? '-'}'),
                  _filterRow('Max size', '${query['max_size'] ?? '-'}'),
                  _filterRow('Cursor', '${query['cursor'] ?? 0}'),
                  _filterRow('Limit', '${query['limit'] ?? _pageSize}'),
                ],
              ),
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(context).pop(),
              child: const Text('Close'),
            ),
          ],
        );
      },
    );
  }

  Future<void> _search({bool preserveSnapshot = false}) async {
    final requestSerial = ++_searchRequestSerial;

    if (!preserveSnapshot) {
      _currentPage = 1;
      _pageCursors
        ..clear()
        ..add(0);
    }

    final baseQuery = preserveSnapshot && _activeSnapshotBaseQuery != null
        ? Map<String, dynamic>.from(_activeSnapshotBaseQuery!)
        : _queryWithoutPagination(_buildQuery());
    final query = _buildQueryWithPage(baseQuery);
    final now = DateTime.now();
    if (_searchNameController.text.trim().isEmpty) {
      _searchNameController.text = _defaultSearchName();
    }
    final searchName = _searchNameController.text.trim();
    final shouldReuseSnapshot = preserveSnapshot && _activeSnapshotId != null;
    final searchId = shouldReuseSnapshot
        ? _activeSnapshotId!
        : 'S-${now.millisecondsSinceEpoch}';
    final createdAt = shouldReuseSnapshot
        ? (_activeSnapshotCreatedAt ?? now)
        : now;

    setState(() {
      _searching = true;
    });

    try {
      final result = await widget.apiClient.searchBlobs(
        accessToken: widget.session.accessToken,
        query: query,
      );

      if (!mounted || requestSerial != _searchRequestSerial) {
        return;
      }

      final blobs = result.blobs;
      var blobIds = blobs.map((item) => item.id).toSet().toList();
      if (preserveSnapshot) {
        final existing = _selectedSnapshot(
          BlobSearchRegistry.instance.snapshots,
        );
        if (existing != null && existing.blobIds.isNotEmpty) {
          blobIds = existing.blobIds;
        }
      }

      await BlobSearchRegistry.instance.upsert(
        BlobSearchSnapshot(
          id: searchId,
          name: searchName,
          createdAt: createdAt,
          blobIds: blobIds,
          resultCount: result.total,
          filters: Map<String, dynamic>.from(query),
        ),
      );

      final currentCursor = _pageCursors[_currentPage - 1];
      final pageCount = result.total <= 0
          ? 1
          : (result.total + _pageSize - 1) ~/ _pageSize;
      final canAdvanceToNextPage =
          _currentPage < pageCount &&
          result.hasMore &&
          result.nextCursor > 0 &&
          result.nextCursor != currentCursor;

      setState(() {
        _hasSearched = true;
        _latestQuery = query;
        _latestResults = blobs;
        _activeSnapshotId = searchId;
        _activeSnapshotCreatedAt = createdAt;
        _activeSnapshotBaseQuery = baseQuery;
        _hasMoreResults = canAdvanceToNextPage;
        _nextCursor = canAdvanceToNextPage ? result.nextCursor : 0;
        _latestTotalMatches = result.total;

        if (canAdvanceToNextPage) {
          if (_pageCursors.length == _currentPage) {
            _pageCursors.add(result.nextCursor);
          } else {
            _pageCursors[_currentPage] = result.nextCursor;
          }
        } else if (_pageCursors.length > _pageCount()) {
          _pageCursors.removeRange(_pageCount(), _pageCursors.length);
        } else if (_pageCursors.length > _currentPage) {
          _pageCursors.removeRange(_currentPage, _pageCursors.length);
        }
      });

      if (mounted) {
        _showActionMessage(
          'Loaded ${blobs.length} result${blobs.length == 1 ? '' : 's'} on page $_currentPage (${result.total} total matches).',
        );
      }

      if (!preserveSnapshot) {
        _hydrateSnapshotWithAllIds(
          snapshotId: searchId,
          searchName: searchName,
          createdAt: createdAt,
          baseQuery: baseQuery,
          firstPage: result,
          filters: Map<String, dynamic>.from(query),
        );
      }
    } catch (error) {
      if (!mounted) {
        return;
      }
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(error.toString())));
    } finally {
      if (mounted) {
        setState(() {
          _searching = false;
        });
      }
    }
  }

  Future<void> _clearSearch() async {
    setState(() {
      _hasSearched = false;
      _currentPage = 1;
      _latestQuery = null;
      _latestResults = null;
      _activeSnapshotId = null;
      _activeSnapshotCreatedAt = null;
      _activeSnapshotBaseQuery = null;
      _hasMoreResults = false;
      _nextCursor = 0;
      _latestTotalMatches = 0;
      _pageCursors
        ..clear()
        ..add(0);
      _searchNameController.text = _defaultSearchName();
    });

    _showActionMessage('Search view cleared.');
  }

  Future<List<int>> _collectAllMatchingBlobIds({
    required Map<String, dynamic> baseQuery,
    required BlobSearchPageResult firstPage,
  }) async {
    final ids = <int>{...firstPage.blobs.map((item) => item.id)};
    var hasMore = firstPage.hasMore;
    var cursor = firstPage.nextCursor;
    final seenCursors = <int>{};
    var pagesFetched = 0;

    while (hasMore) {
      if (cursor <= 0) {
        break;
      }
      if (seenCursors.contains(cursor)) {
        break;
      }
      seenCursors.add(cursor);
      pagesFetched += 1;
      if (pagesFetched > 5000) {
        break;
      }

      final pageQuery = <String, dynamic>{
        ...baseQuery,
        'cursor': cursor,
        'limit': _pageSize,
      };
      final page = await widget.apiClient.searchBlobs(
        accessToken: widget.session.accessToken,
        query: pageQuery,
      );
      ids.addAll(page.blobs.map((item) => item.id));
      if (page.hasMore && page.nextCursor == cursor) {
        break;
      }
      hasMore = page.hasMore;
      cursor = page.nextCursor;
    }

    return ids.toList();
  }

  Future<void> _hydrateSnapshotWithAllIds({
    required String snapshotId,
    required String searchName,
    required DateTime createdAt,
    required Map<String, dynamic> baseQuery,
    required BlobSearchPageResult firstPage,
    required Map<String, dynamic> filters,
  }) async {
    if (!firstPage.hasMore) {
      return;
    }

    final allIds = await _collectAllMatchingBlobIds(
      baseQuery: baseQuery,
      firstPage: firstPage,
    );

    await BlobSearchRegistry.instance.upsert(
      BlobSearchSnapshot(
        id: snapshotId,
        name: searchName,
        createdAt: createdAt,
        blobIds: allIds,
        resultCount: firstPage.total,
        filters: filters,
      ),
    );

    if (mounted && _activeSnapshotId == snapshotId) {
      _showActionMessage('Saved ${allIds.length} matching IDs to history.');
    }
  }

  BlobSearchSnapshot? _selectedSnapshot(List<BlobSearchSnapshot> snapshots) {
    final activeId = _activeSnapshotId;
    if (activeId == null) {
      return null;
    }

    for (final snapshot in snapshots) {
      if (snapshot.id == activeId) {
        return snapshot;
      }
    }
    return null;
  }

  Future<void> _saveSearchName() async {
    final snapshotId = _activeSnapshotId;
    final nextName = _searchNameController.text.trim();

    if (snapshotId == null) {
      if (!mounted) {
        return;
      }
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Select a saved search first, then save the new name.'),
        ),
      );
      return;
    }

    if (nextName.isEmpty) {
      if (!mounted) {
        return;
      }
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Search name cannot be empty.')),
      );
      return;
    }

    await BlobSearchRegistry.instance.rename(snapshotId, nextName);
    if (!mounted) {
      return;
    }
    ScaffoldMessenger.of(
      context,
    ).showSnackBar(const SnackBar(content: Text('Saved search name updated.')));
  }

  Future<void> _chooseSavedSearch() async {
    final snapshots = BlobSearchRegistry.instance.snapshots;
    if (snapshots.isEmpty) {
      if (!mounted) {
        return;
      }
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('No saved searches yet. Run a search first.'),
        ),
      );
      return;
    }

    final selectedId = await showGeneralDialog<String>(
      context: context,
      barrierDismissible: true,
      barrierLabel: 'Close saved searches',
      barrierColor: Colors.black54,
      transitionDuration: const Duration(milliseconds: 220),
      pageBuilder: (context, animation, secondaryAnimation) {
        final screenWidth = MediaQuery.of(context).size.width;
        final panelWidth = screenWidth < 840 ? screenWidth * 0.92 : 460.0;

        return Align(
          alignment: Alignment.centerRight,
          child: SafeArea(
            child: Material(
              color: Theme.of(context).colorScheme.surface,
              elevation: 12,
              borderRadius: const BorderRadius.only(
                topLeft: Radius.circular(16),
                bottomLeft: Radius.circular(16),
              ),
              clipBehavior: Clip.antiAlias,
              child: SizedBox(
                width: panelWidth,
                child: Padding(
                  padding: const EdgeInsets.all(12),
                  child: Column(
                    children: [
                      Row(
                        children: [
                          const Expanded(
                            child: Text(
                              'Select saved search',
                              style: TextStyle(
                                fontWeight: FontWeight.w700,
                                fontSize: 16,
                              ),
                            ),
                          ),
                          IconButton(
                            onPressed: () => Navigator.of(context).pop(),
                            icon: const Icon(Icons.close),
                            tooltip: 'Close',
                          ),
                        ],
                      ),
                      const SizedBox(height: 6),
                      Expanded(
                        child: ListView.separated(
                          itemCount: snapshots.length,
                          separatorBuilder: (_, _) =>
                              const SizedBox(height: 10),
                          itemBuilder: (context, index) {
                            final snapshot = snapshots[index];
                            final selected = snapshot.id == _activeSnapshotId;
                            final local = snapshot.createdAt.toLocal();
                            final month = local.month.toString().padLeft(
                              2,
                              '0',
                            );
                            final day = local.day.toString().padLeft(2, '0');
                            final hour = local.hour.toString().padLeft(2, '0');
                            final minute = local.minute.toString().padLeft(
                              2,
                              '0',
                            );
                            final created =
                                '${local.year}-$month-$day $hour:$minute';

                            return Material(
                              color: selected
                                  ? Colors.lightGreenAccent.withValues(
                                      alpha: 0.12,
                                    )
                                  : Colors.white.withValues(alpha: 0.03),
                              shape: RoundedRectangleBorder(
                                borderRadius: BorderRadius.circular(12),
                                side: BorderSide(
                                  color: selected
                                      ? Colors.lightGreenAccent
                                      : Colors.white24,
                                  width: selected ? 1.4 : 1,
                                ),
                              ),
                              child: InkWell(
                                borderRadius: BorderRadius.circular(12),
                                onTap: () =>
                                    Navigator.of(context).pop(snapshot.id),
                                child: Padding(
                                  padding: const EdgeInsets.all(12),
                                  child: Row(
                                    crossAxisAlignment:
                                        CrossAxisAlignment.start,
                                    children: [
                                      Icon(
                                        selected
                                            ? Icons.check_circle
                                            : Icons.history,
                                        color: selected
                                            ? Colors.lightGreenAccent
                                            : Colors.white70,
                                      ),
                                      const SizedBox(width: 10),
                                      Expanded(
                                        child: Column(
                                          crossAxisAlignment:
                                              CrossAxisAlignment.start,
                                          children: [
                                            Text(
                                              snapshot.name,
                                              maxLines: 1,
                                              overflow: TextOverflow.ellipsis,
                                              style: const TextStyle(
                                                fontWeight: FontWeight.w700,
                                              ),
                                            ),
                                            const SizedBox(height: 4),
                                            Text(
                                              '${snapshot.resultCount} matches • $created',
                                              style: TextStyle(
                                                color: Colors.white.withValues(
                                                  alpha: 0.78,
                                                ),
                                              ),
                                            ),
                                          ],
                                        ),
                                      ),
                                    ],
                                  ),
                                ),
                              ),
                            );
                          },
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        );
      },
      transitionBuilder: (context, animation, secondaryAnimation, child) {
        final offsetTween = Tween<Offset>(
          begin: const Offset(1, 0),
          end: Offset.zero,
        );
        return SlideTransition(
          position: animation.drive(offsetTween),
          child: FadeTransition(opacity: animation, child: child),
        );
      },
    );

    if (selectedId == null) {
      return;
    }

    final selected = snapshots.firstWhere(
      (item) => item.id == selectedId,
      orElse: () => snapshots.first,
    );
    await _resumeSavedSearch(selected);
  }

  Future<void> _refreshCurrentPage() async {
    if (_searching || !_hasSearched) {
      return;
    }
    _showActionMessage('Refreshing page $_currentPage...');
    await _search(preserveSnapshot: true);
  }

  void _showActionMessage(String message) {
    if (!mounted) {
      return;
    }
    ScaffoldMessenger.of(context)
      ..hideCurrentSnackBar()
      ..showSnackBar(
        SnackBar(
          content: Text(message),
          duration: const Duration(milliseconds: 1600),
        ),
      );
  }

  void _setPage(int page) {
    if (page < 1 || page == _currentPage || _searching || !_hasSearched) {
      return;
    }

    final pageCount = _pageCount();
    if (page > pageCount) {
      return;
    }

    final pageIndex = page - 1;
    if (pageIndex >= _pageCursors.length) {
      return;
    }

    setState(() {
      _currentPage = page;
    });
    _showActionMessage('Loading page $page...');
    unawaited(_search(preserveSnapshot: true));
  }

  List<int> _visiblePages() {
    final pageCount = _pageCount();
    if (pageCount <= 1) {
      return const <int>[];
    }

    final visibleCount = _pageCursors.length < pageCount
        ? _pageCursors.length
        : pageCount;
    return List<int>.generate(visibleCount, (index) => index + 1);
  }

  Widget _filterRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 6),
      child: RichText(
        text: TextSpan(
          style: const TextStyle(color: Colors.white),
          children: [
            TextSpan(
              text: '$label: ',
              style: const TextStyle(fontWeight: FontWeight.w700),
            ),
            TextSpan(
              text: value,
              style: TextStyle(color: Colors.white.withValues(alpha: 0.9)),
            ),
          ],
        ),
      ),
    );
  }

  Widget _infoHint(String message) {
    return Tooltip(
      message: message,
      child: Icon(
        Icons.info_outline,
        size: 18,
        color: Colors.white.withValues(alpha: 0.78),
      ),
    );
  }

  Widget _blobDetails(BlobItem blob) {
    return Padding(
      padding: const EdgeInsets.only(top: 10),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('Hash: ${blob.hash}'),
          const SizedBox(height: 6),
          Text('MIME type: ${blob.mimeType}'),
          const SizedBox(height: 6),
          Text('Size: ${_formatBytes(blob.size)}'),
          const SizedBox(height: 10),
          const Text(
            'Locations',
            style: TextStyle(fontWeight: FontWeight.w700),
          ),
          const SizedBox(height: 4),
          if (blob.locations.isEmpty)
            const Text('No locations reported')
          else
            ...blob.locations.map(
              (location) => Padding(
                padding: const EdgeInsets.only(bottom: 4),
                child: SelectableText(location),
              ),
            ),
          const SizedBox(height: 10),
          const Text('Labels', style: TextStyle(fontWeight: FontWeight.w700)),
          const SizedBox(height: 4),
          if (blob.labels.isEmpty)
            const Text('No labels reported')
          else
            ...blob.labels.map(
              (label) => Padding(
                padding: const EdgeInsets.only(bottom: 4),
                child: Text('${label.name} = ${label.value}'),
              ),
            ),
        ],
      ),
    );
  }

  Future<void> _openAddFilterDialog(_FilterKind kind) async {
    final primaryController = TextEditingController();
    var include = true;

    final rule = await showDialog<_FilterRule>(
      context: context,
      builder: (context) {
        return StatefulBuilder(
          builder: (context, setDialogState) {
            var primaryLabel = 'Value';
            var primaryHint = '';

            switch (kind) {
              case _FilterKind.checksums:
                primaryLabel = 'Checksums';
                primaryHint = 'abc123, def456';
                break;
              case _FilterKind.fileTypes:
                primaryLabel = 'File types';
                primaryHint = 'Select one or more options';
                break;
              case _FilterKind.labels:
                primaryLabel = 'Labels and label values';
                primaryHint = 'topic, language=en, owner=team-a';
                break;
              case _FilterKind.pathPattern:
                primaryLabel = 'Path pattern';
                primaryHint = '*.jpg or /docs/*';
                break;
              case _FilterKind.sizeRange:
                primaryLabel = 'Size range';
                primaryHint = 'Choose a default range';
                break;
            }

            final selectedFileTypes = _parseCsv(primaryController.text).toSet();
            final selectedSizePresetId = primaryController.text.trim().isEmpty
                ? null
                : primaryController.text.trim();

            return AlertDialog(
              title: Text('Add ${kind.label} filter'),
              content: SizedBox(
                width: 460,
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    if (kind != _FilterKind.sizeRange)
                      Wrap(
                        spacing: 8,
                        children: [
                          ChoiceChip(
                            label: const Text('Include'),
                            selected: include,
                            onSelected: (_) {
                              setDialogState(() {
                                include = true;
                              });
                            },
                          ),
                          ChoiceChip(
                            label: const Text('Exclude'),
                            selected: !include,
                            onSelected: (_) {
                              setDialogState(() {
                                include = false;
                              });
                            },
                          ),
                        ],
                      ),
                    if (kind != _FilterKind.sizeRange)
                      const SizedBox(height: 12),
                    if (kind == _FilterKind.fileTypes)
                      ConstrainedBox(
                        constraints: const BoxConstraints(maxHeight: 260),
                        child: ListView(
                          shrinkWrap: true,
                          children: _fileTypeOptions.map((item) {
                            final selected = selectedFileTypes.contains(item);
                            return CheckboxListTile(
                              contentPadding: EdgeInsets.zero,
                              title: Text(item),
                              value: selected,
                              onChanged: (next) {
                                final updated = selectedFileTypes.toSet();
                                if (next == true) {
                                  updated.add(item);
                                } else {
                                  updated.remove(item);
                                }
                                primaryController.text = updated.join(',');
                                setDialogState(() {});
                              },
                            );
                          }).toList(),
                        ),
                      )
                    else if (kind == _FilterKind.sizeRange)
                      Column(
                        children: _sizePresets.map((preset) {
                          return RadioListTile<String>(
                            contentPadding: EdgeInsets.zero,
                            title: Text(preset.label),
                            value: preset.id,
                            groupValue: selectedSizePresetId,
                            onChanged: (next) {
                              primaryController.text = next ?? '';
                              setDialogState(() {});
                            },
                          );
                        }).toList(),
                      )
                    else
                      TextField(
                        controller: primaryController,
                        keyboardType: TextInputType.text,
                        decoration: InputDecoration(
                          labelText: primaryLabel,
                          hintText: primaryHint,
                        ),
                      ),
                  ],
                ),
              ),
              actions: [
                TextButton(
                  onPressed: () => Navigator.of(context).pop(),
                  child: const Text('Cancel'),
                ),
                FilledButton(
                  onPressed: () {
                    final primary = primaryController.text.trim();
                    if (primary.isEmpty) {
                      return;
                    }
                    if (kind == _FilterKind.fileTypes &&
                        _parseCsv(primary).isEmpty) {
                      return;
                    }
                    Navigator.of(context).pop(
                      _FilterRule(kind: kind, include: include, value: primary),
                    );
                  },
                  child: const Text('Add'),
                ),
              ],
            );
          },
        );
      },
    );

    if (rule == null) {
      return;
    }

    setState(() {
      _activeFilters.add(rule);
    });
  }

  @override
  Widget build(BuildContext context) {
    final snapshots = BlobSearchRegistry.instance.snapshots;
    final selectedSnapshot = _selectedSnapshot(snapshots);
    final isRenaming = selectedSnapshot != null;

    return Column(
      children: [
        // Search controls
        Card(
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Left side - Search name and action buttons
                Expanded(
                  flex: 2,
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          const Icon(Icons.search, size: 22),
                          const SizedBox(width: 8),
                          const Text(
                            'Blob Search',
                            style: TextStyle(
                              fontSize: 20,
                              fontWeight: FontWeight.w700,
                            ),
                          ),
                        ],
                      ),
                      const SizedBox(height: 12),
                      TextField(
                        controller: _searchNameController,
                        decoration: InputDecoration(
                          labelText: 'Search Name',
                          hintText: 'ex: Weekly docs filter',
                          prefixIcon: const Icon(Icons.edit_note),
                          border: OutlineInputBorder(
                            borderRadius: BorderRadius.circular(8),
                          ),
                        ),
                        style: const TextStyle(fontSize: 14),
                      ),
                      const SizedBox(height: 12),
                      // Status indicator
                      if (isRenaming)
                        Row(
                          children: [
                            Icon(
                              Icons.check_circle,
                              size: 16,
                              color: Colors.lightGreenAccent.withValues(alpha: 0.9),
                            ),
                            const SizedBox(width: 6),
                            Expanded(
                              child: Text(
                                'Active: ${selectedSnapshot!.name}',
                                style: TextStyle(
                                  fontSize: 13,
                                  color: Colors.lightGreenAccent.withValues(alpha: 0.9),
                                  fontWeight: FontWeight.w600,
                                ),
                                overflow: TextOverflow.ellipsis,
                              ),
                            ),
                          ],
                        )
                      else
                        Text(
                          snapshots.isEmpty ? 'No saved searches' : 'New search',
                          style: TextStyle(
                            fontSize: 13,
                            color: Colors.white.withValues(alpha: 0.7),
                          ),
                        ),
                      const SizedBox(height: 12),
                      // Action buttons in a row
                      Wrap(
                        spacing: 8,
                        runSpacing: 8,
                        children: [
                          OutlinedButton.icon(
                            onPressed: _searching ? null : _chooseSavedSearch,
                            style: OutlinedButton.styleFrom(
                              foregroundColor: Colors.white,
                              side: const BorderSide(color: Colors.white54),
                              padding: const EdgeInsets.symmetric(
                                horizontal: 12,
                                vertical: 10,
                              ),
                            ),
                            icon: const Icon(Icons.history, size: 18),
                            label: const Text('History'),
                          ),
                          OutlinedButton.icon(
                            onPressed: _searching || !isRenaming ? null : _saveSearchName,
                            style: OutlinedButton.styleFrom(
                              foregroundColor: Colors.white,
                              side: const BorderSide(color: Colors.white54),
                              padding: const EdgeInsets.symmetric(
                                horizontal: 12,
                                vertical: 10,
                              ),
                            ),
                            icon: const Icon(Icons.save_outlined, size: 18),
                            label: const Text('Save Name'),
                          ),
                          FilledButton.icon(
                            style: FilledButton.styleFrom(
                              backgroundColor: Colors.lightGreenAccent,
                              foregroundColor: Colors.black,
                              padding: const EdgeInsets.symmetric(
                                horizontal: 16,
                                vertical: 10,
                              ),
                              textStyle: const TextStyle(
                                fontWeight: FontWeight.w800,
                                fontSize: 14,
                              ),
                            ),
                            onPressed: _searching
                                ? null
                                : () {
                                    _showActionMessage('Running search...');
                                    unawaited(_search(preserveSnapshot: false));
                                  },
                            icon: _searching
                                ? const SizedBox(
                                    width: 16,
                                    height: 16,
                                    child: CircularProgressIndicator(strokeWidth: 2),
                                  )
                                : const Icon(Icons.search, size: 18),
                            label: Text(_searching ? 'Searching...' : 'Search'),
                          ),
                          OutlinedButton(
                            style: OutlinedButton.styleFrom(
                              foregroundColor: Colors.white,
                              side: const BorderSide(color: Colors.white54),
                              padding: const EdgeInsets.symmetric(
                                horizontal: 12,
                                vertical: 10,
                              ),
                            ),
                            onPressed: _searching ? null : _clearSearch,
                            child: const Text('Clear'),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
                
                const SizedBox(width: 16),
                const VerticalDivider(width: 1),
                const SizedBox(width: 16),
                
                // Right side - Filters
                Expanded(
                  flex: 3,
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text(
                        'Filters',
                        style: TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                      const SizedBox(height: 8),
                      Wrap(
                        spacing: 8,
                        runSpacing: 8,
                        children: [
                          for (final kind in _FilterKind.values)
                            OutlinedButton.icon(
                              onPressed: () => _openAddFilterDialog(kind),
                              style: OutlinedButton.styleFrom(
                                foregroundColor: Colors.white,
                                side: const BorderSide(color: Colors.white54),
                                backgroundColor: Colors.white.withValues(alpha: 0.04),
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 12,
                                  vertical: 10,
                                ),
                              ),
                              icon: const Icon(Icons.add, size: 18),
                              label: Text(kind.label),
                            ),
                        ],
                      ),
                      // Active filters display
                      if (_activeFilters.isNotEmpty) ...[
                        const SizedBox(height: 12),
                        Text(
                          'Active filters (${_activeFilters.length})',
                          style: const TextStyle(
                            fontWeight: FontWeight.w700,
                            fontSize: 13,
                          ),
                        ),
                        const SizedBox(height: 8),
                        Wrap(
                          spacing: 8,
                          runSpacing: 8,
                          children: _activeFilters.asMap().entries.map(
                            (entry) {
                              final filter = entry.value;
                              String label = filter.value;
                              if (filter.kind == _FilterKind.sizeRange) {
                                final preset = _sizePresets
                                    .where((item) => item.id == filter.value)
                                    .cast<_SizePreset?>()
                                    .firstWhere(
                                      (item) => item != null,
                                      orElse: () => null,
                                    );
                                label = preset?.label ?? filter.value;
                              }
                              final prefix = filter.include ? '+' : '-';
                              final filterLabel = filter.kind == _FilterKind.sizeRange
                                  ? label
                                  : '$prefix${filter.kind.label}: $label';
                              return Chip(
                                label: Text(
                                  filterLabel,
                                  style: const TextStyle(fontSize: 12),
                                ),
                                onDeleted: () {
                                  setState(() {
                                    _activeFilters.removeAt(entry.key);
                                  });
                                },
                                deleteIcon: const Icon(Icons.close, size: 16),
                                padding: const EdgeInsets.symmetric(horizontal: 8),
                              );
                            },
                          ).toList(),
                        ),
                      ],
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
        
        const SizedBox(height: 12),
        
        // Results section
        Expanded(
          child: Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Results header with controls
                  Row(
                    children: [
                      const Text(
                        'Search Results',
                        style: TextStyle(
                          fontWeight: FontWeight.w700,
                          fontSize: 16,
                        ),
                      ),
                      const SizedBox(width: 8),
                      if (_latestResults != null && _latestResults!.isNotEmpty) ...[
                        Chip(
                          label: Text('${_latestResults!.length} items'),
                          padding: const EdgeInsets.symmetric(horizontal: 8),
                        ),
                        const SizedBox(width: 8),
                        Chip(
                          label: Text('$_latestTotalMatches total'),
                          padding: const EdgeInsets.symmetric(horizontal: 8),
                        ),
                      ],
                      const Spacer(),
                      if (_hasSearched) ...[
                        OutlinedButton.icon(
                          onPressed: _searching ? null : _refreshCurrentPage,
                          style: OutlinedButton.styleFrom(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 12,
                              vertical: 10,
                            ),
                          ),
                          icon: _searching
                              ? const SizedBox(
                                  width: 14,
                                  height: 14,
                                  child: CircularProgressIndicator(strokeWidth: 1.5),
                                )
                              : const Icon(Icons.refresh, size: 18),
                          label: const Text('Refresh'),
                        ),
                        const SizedBox(width: 8),
                        OutlinedButton.icon(
                          onPressed: _showFiltersDialog,
                          style: OutlinedButton.styleFrom(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 12,
                              vertical: 10,
                            ),
                          ),
                          icon: const Icon(Icons.filter_alt_outlined, size: 18),
                          label: const Text('View Filters'),
                        ),
                      ],
                    ],
                  ),
                  
                  // Page controls
                  if (_hasSearched && _pageCount() > 1) ...[
                    const SizedBox(height: 12),
                    Row(
                      children: [
                        Text(
                          'Page $_currentPage of ${_pageCount()}',
                          style: const TextStyle(
                            fontWeight: FontWeight.w600,
                            fontSize: 13,
                          ),
                        ),
                        const SizedBox(width: 12),
                        OutlinedButton(
                          onPressed: _currentPage > 1
                              ? () => _setPage(_currentPage - 1)
                              : null,
                          style: OutlinedButton.styleFrom(
                            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                          ),
                          child: const Text('Previous'),
                        ),
                        const SizedBox(width: 8),
                        ..._visiblePages().map((page) {
                          final selected = page == _currentPage;
                          if (selected) {
                            return Padding(
                              padding: const EdgeInsets.symmetric(horizontal: 2),
                              child: FilledButton(
                                onPressed: () {},
                                style: FilledButton.styleFrom(
                                  backgroundColor: Colors.white,
                                  foregroundColor: Colors.black,
                                  padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                                  minimumSize: Size.zero,
                                ),
                                child: Text('$page'),
                              ),
                            );
                          }
                          return Padding(
                            padding: const EdgeInsets.symmetric(horizontal: 2),
                            child: OutlinedButton(
                              onPressed: () => _setPage(page),
                              style: OutlinedButton.styleFrom(
                                padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                                minimumSize: Size.zero,
                              ),
                              child: Text('$page'),
                            ),
                          );
                        }),
                        const SizedBox(width: 8),
                        OutlinedButton(
                          onPressed: _hasMoreResults
                              ? () => _setPage(_currentPage + 1)
                              : null,
                          style: OutlinedButton.styleFrom(
                            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                          ),
                          child: const Text('Next'),
                        ),
                      ],
                    ),
                  ],
                  
                  const SizedBox(height: 12),
                  
                  // Results list
                  Expanded(
                    child: _latestResults == null
                        ? const Center(
                            child: Text(
                              'No search has been executed yet.',
                              style: TextStyle(fontSize: 14),
                            ),
                          )
                        : _latestResults!.isEmpty
                        ? const Center(
                            child: Text(
                              'Search completed with 0 results.',
                              style: TextStyle(fontSize: 14),
                            ),
                          )
                        : ListView.separated(
                            itemCount: _latestResults!.length,
                            separatorBuilder: (_, _) =>
                                const SizedBox(height: 8),
                            itemBuilder: (context, index) {
                              final blob = _latestResults![index];
                              return Card(
                                margin: EdgeInsets.zero,
                                child: ExpansionTile(
                                  tilePadding: const EdgeInsets.symmetric(
                                    horizontal: 16,
                                    vertical: 8,
                                  ),
                                  title: Text(
                                    blob.hash,
                                    style: const TextStyle(fontSize: 14),
                                  ),
                                  subtitle: Text(
                                    'ID #${blob.id} • ${blob.mimeType} • ${_formatBytes(blob.size)}',
                                    style: const TextStyle(fontSize: 12),
                                  ),
                                  trailing: Text(
                                    '#${blob.id}',
                                    style: const TextStyle(fontSize: 12),
                                  ),
                                  children: [
                                    Padding(
                                      padding: const EdgeInsets.all(16),
                                      child: _blobDetails(blob),
                                    ),
                                  ],
                                ),
                              );
                            },
                          ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ],
    );
  }
}

enum _FilterKind {
  checksums('Checksums'),
  fileTypes('File types'),
  labels('Labels / Label values'),
  pathPattern('Path pattern'),
  sizeRange('Size range');

  const _FilterKind(this.label);

  final String label;
}

class _FilterRule {
  const _FilterRule({
    required this.kind,
    required this.include,
    required this.value,
  });

  final _FilterKind kind;
  final bool include;
  final String value;
}

class _LabelFilterSplit {
  const _LabelFilterSplit({required this.labels, required this.labelValues});

  final List<String> labels;
  final Map<String, String> labelValues;
}

class _SizePreset {
  const _SizePreset({
    required this.id,
    required this.label,
    required this.min,
    required this.max,
  });

  final String id;
  final String label;
  final int min;
  final int? max;
}