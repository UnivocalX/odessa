import 'dart:convert';
import 'dart:io';

import 'package:file_selector/file_selector.dart';
import 'package:flutter/material.dart';

import '../../models.dart';
import '../../services/api_client.dart';
import '../../services/blob_search_registry.dart';

class BlobHistoryView extends StatefulWidget {
  const BlobHistoryView({
    super.key,
    required this.apiClient,
    required this.session,
  });

  final ApiClient apiClient;
  final AuthSession session;

  @override
  State<BlobHistoryView> createState() => _BlobHistoryViewState();
}

class _BlobHistoryViewState extends State<BlobHistoryView> {
  @override
  void initState() {
    super.initState();
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
    super.dispose();
  }

  String _safeName(String input) {
    final cleaned = input.replaceAll(RegExp(r'[^a-zA-Z0-9._-]'), '_').trim();
    if (cleaned.isEmpty) {
      return 'search_history_ids';
    }
    return cleaned;
  }

  void _showMessage(String message) {
    if (!mounted) {
      return;
    }
    ScaffoldMessenger.of(
      context,
    ).showSnackBar(SnackBar(content: Text(message)));
  }

  Future<void> _downloadIds(BlobSearchSnapshot snapshot) async {
    if (snapshot.blobIds.isEmpty) {
      _showMessage('No IDs to download for this search.');
      return;
    }

    final path = await getSaveLocation(
      suggestedName: '${_safeName(snapshot.name)}_blobs.json',
      acceptedTypeGroups: const [
        XTypeGroup(label: 'JSON files', extensions: ['json']),
      ],
      confirmButtonText: 'Save blobs',
    );
    if (path == null) {
      return;
    }

    final blobs = <Map<String, dynamic>>[];
    for (final blobId in snapshot.blobIds) {
      final blob = await widget.apiClient.getBlobById(
        accessToken: widget.session.accessToken,
        id: blobId,
      );
      blobs.add(blob.toJson());
    }

    final content = const JsonEncoder.withIndent('  ').convert(blobs);
    await File(path.path).writeAsString(content, flush: true);
    _showMessage('Saved ${blobs.length} blob records to ${path.path}');
  }

  Future<void> _showBlobInfo(int blobId) async {
    try {
      final blob = await widget.apiClient.getBlobById(
        accessToken: widget.session.accessToken,
        id: blobId,
      );

      if (!mounted) {
        return;
      }

      await showDialog<void>(
        context: context,
        builder: (context) {
          return AlertDialog(
            title: Text('Blob #${blob.id}'),
            content: SizedBox(
               width: 640,
              child: SingleChildScrollView(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text('Hash: ${blob.hash}'),
                    const SizedBox(height: 6),
                    Text('MIME type: ${blob.mimeType}'),
                    const SizedBox(height: 6),
                    Text('Size: ${blob.size} bytes'),
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
                    const Text(
                      'Labels',
                      style: TextStyle(fontWeight: FontWeight.w700),
                    ),
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
    } catch (error) {
      if (!mounted) {
        return;
      }
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(error.toString())));
    }
  }

  Future<void> _openBlobs(BlobSearchSnapshot snapshot) async {
    await showDialog<void>(
      context: context,
      builder: (context) {
        return AlertDialog(
          title: Text('Blobs in ${snapshot.name}'),
          content: SizedBox(
            width: 640,
            child: snapshot.blobIds.isEmpty
                ? const Text('No blobs found')
                : ListView.separated(
                    shrinkWrap: true,
                    itemCount: snapshot.blobIds.length,
                    separatorBuilder: (_, _) => const Divider(height: 1),
                    itemBuilder: (context, index) {
                      final blobId = snapshot.blobIds[index];
                      return ListTile(
                        dense: true,
                        title: Text('Blob #$blobId'),
                        trailing: IconButton(
                          tooltip: 'Show blob details',
                          onPressed: () => _showBlobInfo(blobId),
                          icon: const Icon(Icons.info_outline),
                        ),
                      );
                    },
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

  Future<void> _deleteSnapshot(BlobSearchSnapshot snapshot) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) {
        return AlertDialog(
          title: const Text('Delete search result?'),
          content: Text(
            'Remove "${snapshot.name}" from search history? This cannot be undone.',
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(context).pop(false),
              child: const Text('Cancel'),
            ),
            FilledButton(
              onPressed: () => Navigator.of(context).pop(true),
              child: const Text('Delete'),
            ),
          ],
        );
      },
    );

    if (confirmed != true) {
      return;
    }

    await BlobSearchRegistry.instance.remove(snapshot.id);
    _showMessage('Deleted search result ${snapshot.name}.');
  }

  String _formatLocalDateTime(DateTime value) {
    final local = value.toLocal();
    final month = local.month.toString().padLeft(2, '0');
    final day = local.day.toString().padLeft(2, '0');
    final hour = local.hour.toString().padLeft(2, '0');
    final minute = local.minute.toString().padLeft(2, '0');
    return '${local.year}-$month-$day $hour:$minute';
  }

  String _filterSummary(Map<String, dynamic> filters) {
    final parts = <String>[];

    final include = (filters['include'] as Map<String, dynamic>?) ?? const {};
    final exclude = (filters['exclude'] as Map<String, dynamic>?) ?? const {};

    if (include.isNotEmpty) {
      parts.add('Include: ${include.keys.join(', ')}');
    }
    if (exclude.isNotEmpty) {
      parts.add('Exclude: ${exclude.keys.join(', ')}');
    }
    if (filters['min_size'] != null) {
      parts.add('Min size: ${filters['min_size']}');
    }
    if (filters['max_size'] != null) {
      parts.add('Max size: ${filters['max_size']}');
    }

    return parts.isEmpty ? 'No filters' : parts.join(' | ');
  }

  @override
  Widget build(BuildContext context) {
    final snapshots = BlobSearchRegistry.instance.snapshots;

    if (snapshots.isEmpty) {
      return const Center(child: Text('No search history yet.'));
    }

    return ListView.separated(
      itemCount: snapshots.length,
      separatorBuilder: (_, _) => const SizedBox(height: 10),
      itemBuilder: (context, index) {
        final item = snapshots[index];
        return Card(
          child: Padding(
            padding: const EdgeInsets.all(12),
            child: ExpansionTile(
              tilePadding: EdgeInsets.zero,
              title: Text(item.name),
              subtitle: Text(
                '${_formatLocalDateTime(item.createdAt)} • ${item.resultCount} matches',
              ),
              trailing: IconButton(
                tooltip: 'Delete search result',
                onPressed: () => _deleteSnapshot(item),
                icon: const Icon(Icons.close),
              ),
              children: [
                Container(
                  width: double.infinity,
                  margin: const EdgeInsets.only(bottom: 8),
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                    borderRadius: BorderRadius.circular(10),
                    color: Colors.white.withValues(alpha: 0.04),
                    border: Border.all(color: Colors.white24),
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Snapshot ID: ${item.id}',
                        style: const TextStyle(fontWeight: FontWeight.w700),
                      ),
                      const SizedBox(height: 4),
                      Text('Stored blobs: ${item.blobIds.length}'),
                      const SizedBox(height: 4),
                      Text(_filterSummary(item.filters)),
                      const SizedBox(height: 10),
                      Wrap(
                        spacing: 8,
                        runSpacing: 8,
                        children: [
                          IconButton(
                            tooltip: 'View blobs',
                            onPressed: () => _openBlobs(item),
                            icon: const Icon(Icons.view_list_outlined),
                          ),
                          FilledButton.icon(
                            onPressed: () => _downloadIds(item),
                            icon: const Icon(Icons.download_outlined),
                            label: const Text('Download blobs JSON'),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        );
      },
    );
  }
}
