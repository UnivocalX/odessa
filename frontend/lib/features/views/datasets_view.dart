import 'dart:io';

import 'package:file_selector/file_selector.dart';
import 'package:flutter/material.dart';
import 'package:s3_dart_lite/s3_dart_lite.dart';

import '../../models.dart';
import '../../services/api_client.dart';
import '../../services/blob_search_registry.dart';
import '../../services/storage_download_config_store.dart';

class DatasetsView extends StatefulWidget {
  const DatasetsView({
    super.key,
    required this.apiClient,
    required this.session,
  });

  final ApiClient apiClient;
  final AuthSession session;

  @override
  State<DatasetsView> createState() => _DatasetsViewState();
}

class _DatasetsViewState extends State<DatasetsView> {
  late Future<List<DatasetItem>> _future;

  @override
  void initState() {
    super.initState();
    _future = _load();
  }

  Future<List<DatasetItem>> _load() {
    return widget.apiClient.getDatasets(widget.session.accessToken);
  }

  Future<void> _refresh() async {
    final next = _load();
    setState(() {
      _future = next;
    });
    await next;
  }

  Future<void> _createDataset() async {
    final nameController = TextEditingController();
    final descriptionController = TextEditingController();
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) {
        return AlertDialog(
          title: const Text('Create Dataset'),
          content: SizedBox(
            width: 420,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                TextField(
                  controller: nameController,
                  decoration: const InputDecoration(labelText: 'Name'),
                ),
                const SizedBox(height: 12),
                TextField(
                  controller: descriptionController,
                  decoration: const InputDecoration(labelText: 'Description'),
                ),
              ],
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(context).pop(false),
              child: const Text('Cancel'),
            ),
            FilledButton(
              onPressed: () => Navigator.of(context).pop(true),
              child: const Text('Create'),
            ),
          ],
        );
      },
    );

    if (confirmed != true) {
      return;
    }

    try {
      await widget.apiClient.createDataset(
        accessToken: widget.session.accessToken,
        name: nameController.text.trim(),
        description: descriptionController.text.trim(),
      );
      if (!mounted) {
        return;
      }
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('Dataset created')));
      await _refresh();
    } catch (error) {
      if (!mounted) {
        return;
      }
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(error.toString())));
    }
  }

  Future<void> _openVersions(DatasetItem dataset) async {
    await showDialog<void>(
      context: context,
      builder: (context) {
        return _DatasetVersionsDialog(
          apiClient: widget.apiClient,
          session: widget.session,
          dataset: dataset,
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Align(
          alignment: Alignment.centerRight,
          child: FilledButton.icon(
            onPressed: _createDataset,
            icon: const Icon(Icons.add),
            label: const Text('Create dataset'),
          ),
        ),
        const SizedBox(height: 12),
        Expanded(
          child: FutureBuilder<List<DatasetItem>>(
            future: _future,
            builder: (context, snapshot) {
              if (snapshot.connectionState == ConnectionState.waiting) {
                return const Center(child: CircularProgressIndicator());
              }

              if (snapshot.hasError) {
                return _ErrorPane(
                  message: snapshot.error.toString(),
                  onRetry: _refresh,
                );
              }

              final datasets = snapshot.data ?? const <DatasetItem>[];
              if (datasets.isEmpty) {
                return const _EmptyPane(title: 'No datasets found');
              }

              return RefreshIndicator(
                onRefresh: _refresh,
                child: ListView.separated(
                  itemCount: datasets.length,
                  separatorBuilder: (_, _) => const SizedBox(height: 10),
                  itemBuilder: (context, index) {
                    final item = datasets[index];
                    return Card(
                      child: ListTile(
                        title: Text(item.name),
                        subtitle: Text(
                          item.description.isEmpty
                              ? 'No description'
                              : item.description,
                        ),
                        trailing: Wrap(
                          spacing: 8,
                          children: [
                            OutlinedButton(
                              onPressed: () => _openVersions(item),
                              child: const Text('Versions'),
                            ),
                            Text('#${item.id}'),
                          ],
                        ),
                      ),
                    );
                  },
                ),
              );
            },
          ),
        ),
      ],
    );
  }
}

class _DatasetVersionsDialog extends StatefulWidget {
  const _DatasetVersionsDialog({
    required this.apiClient,
    required this.session,
    required this.dataset,
  });

  final ApiClient apiClient;
  final AuthSession session;
  final DatasetItem dataset;

  @override
  State<_DatasetVersionsDialog> createState() => _DatasetVersionsDialogState();
}

class _DatasetVersionsDialogState extends State<_DatasetVersionsDialog> {
  late Future<List<DatasetVersionItem>> _future;
  final StorageDownloadConfigStore _storageConfigStore =
      StorageDownloadConfigStore();

  String _formatLocalDateTime(String value) {
    final parsed = DateTime.tryParse(value);
    if (parsed == null) {
      return value.isEmpty ? 'Unknown' : value;
    }

    final local = parsed.toLocal();
    final month = local.month.toString().padLeft(2, '0');
    final day = local.day.toString().padLeft(2, '0');
    final hour = local.hour.toString().padLeft(2, '0');
    final minute = local.minute.toString().padLeft(2, '0');
    return '${local.year}-$month-$day $hour:$minute';
  }

  Client? _buildS3Client(StorageDownloadConfig config) {
    final endpointText = config.s3ApiEndpoint.trim();
    final accessKey = config.accessKey.trim();
    final secretKey = config.secretKey.trim();

    if (endpointText.isEmpty || accessKey.isEmpty || secretKey.isEmpty) {
      return null;
    }

    final endpointUri = Uri.tryParse(endpointText);
    final endpointHost = endpointUri != null && endpointUri.hasScheme && endpointUri.host.isNotEmpty
        ? endpointUri.host
        : endpointText.replaceFirst(RegExp(r'^https?://'), '');
    if (endpointHost.isEmpty) {
      return null;
    }

    final useSSL = endpointUri?.scheme == 'https';
    final port = endpointUri != null && endpointUri.hasPort ? endpointUri.port : null;

    return Client(
      ClientOptions(
        endPoint: endpointHost,
        region: 'us-east-1',
        accessKey: accessKey,
        secretKey: secretKey,
        useSSL: useSSL,
        pathStyle: true,
        port: port,
      ),
    );
  }

  ({String bucket, String key})? _parseS3Location(String location) {
    final uri = Uri.tryParse(location);
    if (uri == null || uri.scheme != 's3' || uri.host.isEmpty) {
      return null;
    }

    final key = uri.pathSegments.join('/');
    if (key.isEmpty) {
      return null;
    }

    return (bucket: uri.host, key: key);
  }

  String? _resolveDownloadUrl(BlobItem blob, StorageDownloadConfig config) {
    final template = config.downloadUrlTemplate.trim();
    if (template.isNotEmpty) {
      final preferredLocation = blob.locations.isEmpty
          ? ''
          : blob.locations.first;
      return _applyTemplateOrQuery(
        template,
        location: preferredLocation,
        hash: blob.hash,
      );
    }

    final httpLocation = blob.locations
        .where(
          (item) => item.startsWith('http://') || item.startsWith('https://'),
        )
        .cast<String?>()
        .firstWhere((item) => item != null, orElse: () => null);

    if (httpLocation != null) {
      return httpLocation;
    }

    final s3Location = blob.locations
        .where((item) => item.startsWith('s3://'))
        .cast<String?>()
        .firstWhere((item) => item != null, orElse: () => null);
    if (s3Location != null) {
      final endpoint = config.s3ApiEndpoint.trim();
      if (endpoint.isNotEmpty) {
        return _applyTemplateOrQuery(
          endpoint,
          location: s3Location,
          hash: blob.hash,
        );
      }
    }

    return null;
  }

  String _applyTemplateOrQuery(
    String value, {
    required String location,
    required String hash,
  }) {
    if (value.contains('{location}') || value.contains('{hash}')) {
      return value
          .replaceAll('{location}', Uri.encodeComponent(location))
          .replaceAll('{hash}', Uri.encodeComponent(hash));
    }

    final uri = Uri.tryParse(value);
    if (uri == null || !uri.hasScheme || uri.host.isEmpty) {
      return value;
    }

    final query = Map<String, String>.from(uri.queryParameters)
      ..putIfAbsent('uri', () => location)
      ..putIfAbsent('hash', () => hash);
    return uri.replace(queryParameters: query).toString();
  }

  Map<String, String> _downloadHeaders(StorageDownloadConfig config) {
    final headers = <String, String>{};
    if (config.bearerToken.trim().isNotEmpty) {
      headers['Authorization'] = 'Bearer ${config.bearerToken.trim()}';
    }
    if (config.customHeaderName.trim().isNotEmpty) {
      headers[config.customHeaderName.trim()] = config.customHeaderValue.trim();
    }
    return headers;
  }

  String _joinPath(String dir, String name) {
    if (dir.endsWith(Platform.pathSeparator)) {
      return '$dir$name';
    }
    return '$dir${Platform.pathSeparator}$name';
  }

  String _safeFileName(String hash, String mimeType) {
    final cleaned = hash.replaceAll(RegExp(r'[^a-zA-Z0-9._-]'), '_');
    if (mimeType.contains('json')) {
      return '$cleaned.json';
    }
    if (mimeType.contains('plain')) {
      return '$cleaned.txt';
    }
    if (mimeType.contains('pdf')) {
      return '$cleaned.pdf';
    }
    if (mimeType.contains('png')) {
      return '$cleaned.png';
    }
    if (mimeType.contains('jpeg')) {
      return '$cleaned.jpg';
    }
    return '$cleaned.bin';
  }

  Future<void> _showBlobInfo(BlobItem blob) async {
    try {
      final fullBlob = await widget.apiClient.getBlobByHash(
        accessToken: widget.session.accessToken,
        hash: blob.hash,
      );

      if (!mounted) {
        return;
      }

      await showDialog<void>(
        context: context,
        builder: (context) {
          return AlertDialog(
            title: Text('Blob #${fullBlob.id}'),
            content: SizedBox(
              width: 640,
              child: SingleChildScrollView(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text('Hash: ${fullBlob.hash}'),
                    const SizedBox(height: 6),
                    Text('MIME type: ${fullBlob.mimeType}'),
                    const SizedBox(height: 6),
                    Text('Size: ${fullBlob.size} bytes'),
                    const SizedBox(height: 10),
                    const Text(
                      'Locations',
                      style: TextStyle(fontWeight: FontWeight.w700),
                    ),
                    const SizedBox(height: 4),
                    if (fullBlob.locations.isEmpty)
                      const Text('No locations reported')
                    else
                      ...fullBlob.locations.map(
                        (location) => Padding(
                          padding: const EdgeInsets.only(bottom: 4),
                          child: SelectableText(location),
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

  Future<_VersionDownloadResult> _downloadBlobToFolder({
    required BlobItem blob,
    required String folderPath,
    required StorageDownloadConfig config,
    required Client? s3Client,
  }) async {
    final filePath = _joinPath(
      folderPath,
      _safeFileName(blob.hash, blob.mimeType),
    );
    final file = File(filePath);

    final s3Location = blob.locations
        .where((item) => item.startsWith('s3://'))
        .cast<String?>()
        .firstWhere((item) => item != null, orElse: () => null);
    if (s3Location != null) {
      final parsed = _parseS3Location(s3Location);
      if (parsed == null || s3Client == null) {
        return _VersionDownloadResult(
          hash: blob.hash,
          status: _VersionDownloadStatus.skipped,
          message:
              'Configure the storage endpoint, access key, and secret key in Settings to download S3 objects directly.',
          filePath: filePath,
        );
      }

      try {
        final response = await s3Client.getObject(
          parsed.key,
          bucketName: parsed.bucket,
        );
        await file.writeAsBytes(response.bodyBytes, flush: true);
        return _VersionDownloadResult(
          hash: blob.hash,
          status: _VersionDownloadStatus.saved,
          message: 'Downloaded from s3://${parsed.bucket}/${parsed.key}',
          filePath: filePath,
        );
      } catch (error) {
        return _VersionDownloadResult(
          hash: blob.hash,
          status: _VersionDownloadStatus.failed,
          message: '$error ($s3Location)',
          filePath: filePath,
        );
      }
    }

    final location = _resolveDownloadUrl(blob, config);

    if (location == null) {
      return _VersionDownloadResult(
        hash: blob.hash,
        status: _VersionDownloadStatus.skipped,
        message:
            'No downloadable URL. Add a public download URL template in Settings for non-S3 locations.',
        filePath: null,
      );
    }

    try {
      final client = HttpClient();
      try {
        final request = await client.getUrl(Uri.parse(location));
        final headers = _downloadHeaders(config);
        headers.forEach((key, value) {
          request.headers.set(key, value);
        });
        final response = await request.close();
        if (response.statusCode < 200 || response.statusCode >= 300) {
          return _VersionDownloadResult(
            hash: blob.hash,
            status: _VersionDownloadStatus.failed,
            message: 'HTTP ${response.statusCode} from $location',
            filePath: filePath,
          );
        }
        final bytes = await response.fold<List<int>>(<int>[], (acc, chunk) {
          acc.addAll(chunk);
          return acc;
        });
        await file.writeAsBytes(bytes, flush: true);
        return _VersionDownloadResult(
          hash: blob.hash,
          status: _VersionDownloadStatus.saved,
          message: 'Saved from $location',
          filePath: filePath,
        );
      } finally {
        client.close(force: true);
      }
    } catch (error) {
      return _VersionDownloadResult(
        hash: blob.hash,
        status: _VersionDownloadStatus.failed,
        message: '$error ($location)',
        filePath: filePath,
      );
    }
  }

  Future<List<BlobItem>> _loadAllVersionBlobs(int versionId) async {
    final items = <BlobItem>[];
    var cursor = 0;

    while (true) {
      final page = await widget.apiClient.getDatasetVersionBlobs(
        accessToken: widget.session.accessToken,
        datasetId: widget.dataset.id,
        versionId: versionId,
        cursor: cursor,
        limit: 100,
      );

      items.addAll(page.blobs);
      if (!page.hasMore) {
        break;
      }
      cursor = page.nextCursor;
      if (cursor <= 0) {
        break;
      }
    }

    return items;
  }

  Future<void> _downloadVersionFiles(DatasetVersionItem version) async {
    final folderPath = await getDirectoryPath(
      confirmButtonText: 'Select destination folder',
      initialDirectory: Directory.current.path,
    );

    if (folderPath == null || folderPath.trim().isEmpty) {
      return;
    }

    try {
      final config = await _storageConfigStore.load();
      final blobs = await _loadAllVersionBlobs(version.id);
      final s3Client = _buildS3Client(config);

      if (!mounted) {
        s3Client?.close();
        return;
      }

      if (blobs.isEmpty) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('No blobs found for this version')),
        );
        return;
      }

      final results = <_VersionDownloadResult>[];
      try {
        for (final blob in blobs) {
          results.add(
            await _downloadBlobToFolder(
              blob: blob,
              folderPath: folderPath,
              config: config,
              s3Client: s3Client,
            ),
          );
        }
      } finally {
        s3Client?.close();
      }

      if (!mounted) {
        return;
      }

      await showDialog<void>(
        context: context,
        builder: (context) {
          return AlertDialog(
            title: Text('Downloaded files for version ${version.id}'),
            content: SizedBox(
              width: 760,
              height: 420,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text('Folder: $folderPath'),
                  const SizedBox(height: 10),
                  Expanded(
                    child: ListView.separated(
                      itemCount: results.length,
                      separatorBuilder: (_, _) => const Divider(height: 1),
                      itemBuilder: (context, index) {
                        final item = results[index];
                        final color = switch (item.status) {
                          _VersionDownloadStatus.saved => Colors.greenAccent,
                          _VersionDownloadStatus.skipped => Colors.orangeAccent,
                          _VersionDownloadStatus.failed => Colors.redAccent,
                        };

                        return ListTile(
                          dense: true,
                          title: Text(item.hash),
                          subtitle: Text(
                            item.filePath == null
                                ? item.message
                                : '${item.filePath}\n${item.message}',
                          ),
                          isThreeLine: true,
                          trailing: Text(
                            item.status.name,
                            style: TextStyle(
                              color: color,
                              fontWeight: FontWeight.w700,
                            ),
                          ),
                        );
                      },
                    ),
                  ),
                ],
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
      ).showSnackBar(SnackBar(content: Text('$error')));
    }
  }

  @override
  void initState() {
    super.initState();
    _future = _load();
  }

  Future<List<DatasetVersionItem>> _load() {
    return widget.apiClient.getDatasetVersions(
      accessToken: widget.session.accessToken,
      datasetId: widget.dataset.id,
    );
  }

  Future<void> _refresh() async {
    final next = _load();
    setState(() {
      _future = next;
    });
    await next;
  }

  Future<void> _createVersion() async {
    final commitController = TextEditingController();
    final snapshots = BlobSearchRegistry.instance.snapshots;
    String? selectedSearchId;

    if (snapshots.isNotEmpty) {
      selectedSearchId = snapshots.first.id;
    }

    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) {
        return StatefulBuilder(
          builder: (context, setDialogState) {
            final selected = selectedSearchId == null
                ? null
                : snapshots
                      .where((item) => item.id == selectedSearchId)
                      .cast<BlobSearchSnapshot?>()
                      .firstWhere((item) => item != null, orElse: () => null);

            return AlertDialog(
              title: const Text('Create Dataset Version'),
              content: SizedBox(
                width: 500,
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    TextField(
                      controller: commitController,
                      decoration: const InputDecoration(labelText: 'Commit'),
                    ),
                    const SizedBox(height: 12),
                    DropdownButtonFormField<String>(
                      initialValue: selectedSearchId,
                      decoration: const InputDecoration(
                        labelText: 'Search result snapshot',
                      ),
                      items: snapshots
                          .map(
                            (item) => DropdownMenuItem<String>(
                              value: item.id,
                              child: Text(
                                '${item.name} (${item.resultCount} results)',
                                overflow: TextOverflow.ellipsis,
                              ),
                            ),
                          )
                          .toList(),
                      onChanged: snapshots.isEmpty
                          ? null
                          : (value) {
                              setDialogState(() {
                                selectedSearchId = value;
                              });
                            },
                    ),
                    const SizedBox(height: 10),
                    if (snapshots.isEmpty)
                      const Text(
                        'No saved blob searches yet. Run a search in Blobs first.',
                      )
                    else if (selected != null)
                      Text(
                        'Selected: ${selected.id} • ${selected.resultCount} blobs',
                      ),
                  ],
                ),
              ),
              actions: [
                TextButton(
                  onPressed: () => Navigator.of(context).pop(false),
                  child: const Text('Cancel'),
                ),
                FilledButton(
                  onPressed: snapshots.isEmpty || selectedSearchId == null
                      ? null
                      : () => Navigator.of(context).pop(true),
                  child: const Text('Create'),
                ),
              ],
            );
          },
        );
      },
    );

    if (confirmed != true) {
      return;
    }

    final selected = selectedSearchId == null
        ? null
        : snapshots
              .where((item) => item.id == selectedSearchId)
              .cast<BlobSearchSnapshot?>()
              .firstWhere((item) => item != null, orElse: () => null);
    final blobIds = selected?.blobIds ?? const <int>[];

    if (blobIds.isEmpty) {
      if (!mounted) {
        return;
      }
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Please select a saved search that has results'),
        ),
      );
      return;
    }

    try {
      await widget.apiClient.createDatasetVersion(
        accessToken: widget.session.accessToken,
        datasetId: widget.dataset.id,
        commit: commitController.text.trim(),
        blobIds: blobIds,
      );
      if (!mounted) {
        return;
      }
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('Version created')));
      await _refresh();
    } catch (error) {
      if (!mounted) {
        return;
      }
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(error.toString())));
    }
  }

  Future<void> _openVersionBlobs(DatasetVersionItem version) async {
    try {
      final blobs = await _loadAllVersionBlobs(version.id);
      if (!mounted) {
        return;
      }
      await showDialog<void>(
        context: context,
        builder: (context) {
          return AlertDialog(
            title: Text('Blobs in version ${version.id}'),
            content: SizedBox(
              width: 640,
              child: blobs.isEmpty
                  ? const Text('No blobs found')
                  : ListView.separated(
                      shrinkWrap: true,
                      itemCount: blobs.length,
                      separatorBuilder: (_, _) => const Divider(height: 1),
                      itemBuilder: (context, index) {
                        final blob = blobs[index];
                        return ListTile(
                          dense: true,
                          title: Text(blob.hash),
                          trailing: IconButton(
                            tooltip: 'Show blob details',
                            onPressed: () => _showBlobInfo(blob),
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
    } catch (error) {
      if (!mounted) {
        return;
      }
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(error.toString())));
    }
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Text('Versions: ${widget.dataset.name}'),
      content: SizedBox(
        width: 760,
        child: FutureBuilder<List<DatasetVersionItem>>(
          future: _future,
          builder: (context, snapshot) {
            if (snapshot.connectionState == ConnectionState.waiting) {
              return const Center(child: CircularProgressIndicator());
            }
            if (snapshot.hasError) {
              return Text(snapshot.error.toString());
            }

            final versions = snapshot.data ?? const <DatasetVersionItem>[];
            if (versions.isEmpty) {
              return const Text('No versions yet');
            }

            return ListView.separated(
              shrinkWrap: true,
              itemCount: versions.length,
              separatorBuilder: (_, _) => const Divider(height: 1),
              itemBuilder: (context, index) {
                final version = versions[index];
                return ListTile(
                  title: Text('Version #${version.id}'),
                  subtitle: Text(
                    'Commit: ${version.commit.isEmpty ? '(none)' : version.commit}\n'
                    'Created: ${_formatLocalDateTime(version.createdAt)}\n'
                    'Blob count: ${version.blobIds.length}',
                  ),
                  trailing: Wrap(
                    spacing: 8,
                    children: [
                      OutlinedButton(
                        onPressed: () => _openVersionBlobs(version),
                        child: const Text('Blobs'),
                      ),
                      OutlinedButton.icon(
                        onPressed: () => _downloadVersionFiles(version),
                        icon: const Icon(Icons.download),
                        label: const Text('Download files'),
                      ),
                    ],
                  ),
                );
              },
            );
          },
        ),
      ),
      actions: [
        FilledButton.icon(
          onPressed: _createVersion,
          icon: const Icon(Icons.add),
          label: const Text('Create version'),
        ),
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: const Text('Close'),
        ),
      ],
    );
  }
}

enum _VersionDownloadStatus { saved, skipped, failed }

class _VersionDownloadResult {
  const _VersionDownloadResult({
    required this.hash,
    required this.status,
    required this.message,
    required this.filePath,
  });

  final String hash;
  final _VersionDownloadStatus status;
  final String message;
  final String? filePath;
}

class _ErrorPane extends StatelessWidget {
  const _ErrorPane({required this.message, required this.onRetry});

  final String message;
  final Future<void> Function() onRetry;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            message,
            textAlign: TextAlign.center,
            style: const TextStyle(color: Colors.red),
          ),
          const SizedBox(height: 12),
          FilledButton(onPressed: onRetry, child: const Text('Retry')),
        ],
      ),
    );
  }
}

class _EmptyPane extends StatelessWidget {
  const _EmptyPane({required this.title});

  final String title;

  @override
  Widget build(BuildContext context) {
    return Center(child: Text(title));
  }
}
