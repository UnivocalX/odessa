import 'package:flutter/material.dart';

import '../../models.dart';
import '../../services/api_client.dart';

class LabelsView extends StatefulWidget {
  const LabelsView({super.key, required this.apiClient, required this.session});

  final ApiClient apiClient;
  final AuthSession session;

  @override
  State<LabelsView> createState() => _LabelsViewState();
}

class _LabelsViewState extends State<LabelsView> {
  late Future<List<LabelItem>> _future;

  Widget _metaChip(IconData icon, String value) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(999),
        color: Colors.white.withValues(alpha: 0.05),
        border: Border.all(color: Colors.white24),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [Icon(icon, size: 14), const SizedBox(width: 6), Text(value)],
      ),
    );
  }

  @override
  void initState() {
    super.initState();
    _future = _load();
  }

  Future<List<LabelItem>> _load() {
    return widget.apiClient.getLabels(widget.session.accessToken);
  }

  Future<void> _refresh() async {
    final next = _load();
    setState(() {
      _future = next;
    });
    await next;
  }

  Future<void> _createLabel() async {
    final nameController = TextEditingController();
    final descriptionController = TextEditingController();

    final result = await showDialog<bool>(
      context: context,
      builder: (context) {
        return AlertDialog(
          title: const Text('Create Label'),
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

    if (result != true) {
      return;
    }

    try {
      await widget.apiClient.createLabel(
        accessToken: widget.session.accessToken,
        name: nameController.text.trim(),
        description: descriptionController.text.trim(),
      );
      if (!mounted) {
        return;
      }
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('Label created')));
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

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Align(
          alignment: Alignment.centerRight,
          child: FilledButton.icon(
            onPressed: _createLabel,
            icon: const Icon(Icons.add),
            label: const Text('Create label'),
          ),
        ),
        const SizedBox(height: 12),
        Expanded(
          child: FutureBuilder<List<LabelItem>>(
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

              final labels = snapshot.data ?? const <LabelItem>[];
              if (labels.isEmpty) {
                return const Center(child: Text('No labels found'));
              }

              return RefreshIndicator(
                onRefresh: _refresh,
                child: ListView.separated(
                  itemCount: labels.length,
                  separatorBuilder: (_, _) => const SizedBox(height: 10),
                  itemBuilder: (context, index) {
                    final item = labels[index];
                    final description = item.description.isEmpty
                        ? 'No description'
                        : item.description;
                    final initial = item.name.isEmpty
                        ? '?'
                        : item.name.characters.first.toUpperCase();

                    return Card(
                      child: Padding(
                        padding: const EdgeInsets.all(14),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Row(
                              children: [
                                Container(
                                  width: 38,
                                  height: 38,
                                  decoration: BoxDecoration(
                                    borderRadius: BorderRadius.circular(10),
                                    color: Colors.white.withValues(alpha: 0.08),
                                    border: Border.all(color: Colors.white24),
                                  ),
                                  child: Center(
                                    child: Text(
                                      initial,
                                      style: const TextStyle(
                                        fontWeight: FontWeight.w700,
                                        fontSize: 16,
                                      ),
                                    ),
                                  ),
                                ),
                                const SizedBox(width: 12),
                                Expanded(
                                  child: Text(
                                    item.name,
                                    style: const TextStyle(
                                      fontWeight: FontWeight.w700,
                                      fontSize: 16,
                                    ),
                                  ),
                                ),
                                _metaChip(
                                  Icons.tag_outlined,
                                  'Label #${item.id}',
                                ),
                              ],
                            ),
                            const SizedBox(height: 10),
                            Container(
                              width: double.infinity,
                              padding: const EdgeInsets.all(10),
                              decoration: BoxDecoration(
                                borderRadius: BorderRadius.circular(12),
                                color: Colors.white.withValues(alpha: 0.04),
                                border: Border.all(color: Colors.white24),
                              ),
                              child: Text(description),
                            ),
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
