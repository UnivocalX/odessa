import 'package:flutter/material.dart';

import '../../models.dart';
import '../../services/api_client.dart';

class OriginsView extends StatefulWidget {
  const OriginsView({
    super.key,
    required this.apiClient,
    required this.session,
  });

  final ApiClient apiClient;
  final AuthSession session;

  @override
  State<OriginsView> createState() => _OriginsViewState();
}

class _OriginsViewState extends State<OriginsView> {
  late Future<List<OriginItem>> _future;
  final Map<int, bool> _expanded = {};
  final Map<int, List<ScanOriginItem>> _scanHistoryByOrigin = {};

  @override
  void initState() {
    super.initState();
    _future = _load();
  }

  Future<List<OriginItem>> _load() {
    return widget.apiClient.getOrigins(widget.session.accessToken);
  }

  Future<void> _refresh() async {
    final next = _load();
    setState(() {
      _future = next;
    });
    await next;
  }

  void _upsertScan(ScanOriginItem scan) {
    final list = _scanHistoryByOrigin.putIfAbsent(scan.originId, () => []);
    final existingIndex = list.indexWhere((item) => item.id == scan.id);
    if (existingIndex >= 0) {
      list[existingIndex] = scan;
    } else {
      list.insert(0, scan);
    }
  }

  Future<void> _fetchLatestScan(int originId) async {
    try {
      final scan = await widget.apiClient.getOriginScan(
        accessToken: widget.session.accessToken,
        originId: originId,
      );

      if (!mounted) {
        return;
      }

      setState(() {
        _upsertScan(scan);
      });
    } catch (_) {
      // Ignore not found errors for origins that were never scanned.
    }
  }

  Future<Map<String, dynamic>?> _openRuleBuilder({
    Map<String, dynamic>? seed,
  }) async {
    final patterns = <_RulePatternDraft>[];

    if (seed != null && seed.isNotEmpty) {
      for (final entry in seed.entries) {
        final assignments = <_LabelAssignmentDraft>[];
        if (entry.value is List) {
          for (final assignment in entry.value as List<dynamic>) {
            if (assignment is Map) {
              assignments.add(
                _LabelAssignmentDraft(
                  label: '${assignment['label'] ?? ''}',
                  value: '${assignment['value'] ?? ''}',
                ),
              );
            }
          }
        }
        patterns.add(
          _RulePatternDraft(pattern: entry.key, assignments: assignments),
        );
      }
    }

    if (patterns.isEmpty) {
      patterns.add(_RulePatternDraft(pattern: '*'));
    }

    return showDialog<Map<String, dynamic>>(
      context: context,
      builder: (context) {
        return StatefulBuilder(
          builder: (context, setDialogState) {
            void addPattern() {
              setDialogState(() {
                patterns.add(_RulePatternDraft(pattern: ''));
              });
            }

            void removePattern(int index) {
              setDialogState(() {
                patterns.removeAt(index);
              });
            }

            void addAssignment(int pIndex) {
              setDialogState(() {
                patterns[pIndex].assignments.add(
                  _LabelAssignmentDraft(label: '', value: ''),
                );
              });
            }

            void removeAssignment(int pIndex, int aIndex) {
              setDialogState(() {
                patterns[pIndex].assignments.removeAt(aIndex);
              });
            }

            Map<String, dynamic> toRules() {
              final rules = <String, dynamic>{};
              for (final pattern in patterns) {
                final key = pattern.pattern.trim();
                if (key.isEmpty) {
                  continue;
                }

                final assignments = pattern.assignments
                    .where((assignment) => assignment.label.trim().isNotEmpty)
                    .map(
                      (assignment) => {
                        'label': assignment.label.trim(),
                        'value': assignment.value.trim(),
                      },
                    )
                    .toList();

                if (assignments.isNotEmpty) {
                  rules[key] = assignments;
                }
              }
              return rules;
            }

            return AlertDialog(
              title: const Text('Rule Builder'),
              content: SizedBox(
                width: 800,
                child: ListView.separated(
                  shrinkWrap: true,
                  itemCount: patterns.length,
                  separatorBuilder: (_, _) => const SizedBox(height: 12),
                  itemBuilder: (context, pIndex) {
                    final pattern = patterns[pIndex];
                    return Card(
                      child: Padding(
                        padding: const EdgeInsets.all(12),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Row(
                              children: [
                                Expanded(
                                  child: TextField(
                                    controller: TextEditingController(
                                      text: pattern.pattern,
                                    ),
                                    onChanged: (value) {
                                      pattern.pattern = value;
                                    },
                                    decoration: const InputDecoration(
                                      labelText: 'Pattern (e.g. * or *.jpg)',
                                    ),
                                  ),
                                ),
                                const SizedBox(width: 8),
                                IconButton(
                                  tooltip: 'Remove pattern',
                                  onPressed: () => removePattern(pIndex),
                                  icon: const Icon(Icons.delete_outline),
                                ),
                              ],
                            ),
                            const SizedBox(height: 10),
                            const Text(
                              'Assignments',
                              style: TextStyle(fontWeight: FontWeight.w700),
                            ),
                            const SizedBox(height: 8),
                            if (pattern.assignments.isEmpty)
                              const Text('No assignments yet')
                            else
                              ListView.separated(
                                shrinkWrap: true,
                                physics: const NeverScrollableScrollPhysics(),
                                itemCount: pattern.assignments.length,
                                separatorBuilder: (_, _) =>
                                    const SizedBox(height: 8),
                                itemBuilder: (context, aIndex) {
                                  final assignment =
                                      pattern.assignments[aIndex];
                                  return Row(
                                    children: [
                                      Expanded(
                                        child: TextField(
                                          controller: TextEditingController(
                                            text: assignment.label,
                                          ),
                                          onChanged: (value) {
                                            assignment.label = value;
                                          },
                                          decoration: const InputDecoration(
                                            labelText: 'Label',
                                          ),
                                        ),
                                      ),
                                      const SizedBox(width: 8),
                                      Expanded(
                                        child: TextField(
                                          controller: TextEditingController(
                                            text: assignment.value,
                                          ),
                                          onChanged: (value) {
                                            assignment.value = value;
                                          },
                                          decoration: const InputDecoration(
                                            labelText: 'Value',
                                          ),
                                        ),
                                      ),
                                      const SizedBox(width: 8),
                                      IconButton(
                                        tooltip: 'Remove assignment',
                                        onPressed: () =>
                                            removeAssignment(pIndex, aIndex),
                                        icon: const Icon(
                                          Icons.remove_circle_outline,
                                        ),
                                      ),
                                    ],
                                  );
                                },
                              ),
                            const SizedBox(height: 8),
                            TextButton.icon(
                              onPressed: () => addAssignment(pIndex),
                              icon: const Icon(Icons.add),
                              label: const Text('Add assignment'),
                            ),
                          ],
                        ),
                      ),
                    );
                  },
                ),
              ),
              actions: [
                TextButton.icon(
                  onPressed: addPattern,
                  icon: const Icon(Icons.add),
                  label: const Text('Add pattern'),
                ),
                TextButton(
                  onPressed: () => Navigator.of(context).pop(),
                  child: const Text('Cancel'),
                ),
                FilledButton(
                  onPressed: () => Navigator.of(context).pop(toRules()),
                  child: const Text('Save rules'),
                ),
              ],
            );
          },
        );
      },
    );
  }

  Future<void> _createOrigin() async {
    final uriController = TextEditingController();
    Map<String, dynamic> rules = {};

    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) {
        return StatefulBuilder(
          builder: (context, setDialogState) {
            return AlertDialog(
              title: const Text('Create Origin'),
              content: SizedBox(
                width: 560,
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    TextField(
                      controller: uriController,
                      decoration: const InputDecoration(
                        labelText: 'Origin URI',
                      ),
                    ),
                    const SizedBox(height: 12),
                    Text('Configured rule patterns: ${rules.length}'),
                    const SizedBox(height: 8),
                    OutlinedButton.icon(
                      onPressed: () async {
                        final built = await _openRuleBuilder(seed: rules);
                        if (built == null) {
                          return;
                        }
                        setDialogState(() {
                          rules = built;
                        });
                      },
                      icon: const Icon(Icons.rule),
                      label: const Text('Edit rules'),
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
      },
    );

    if (confirmed != true) {
      return;
    }

    try {
      final message = await widget.apiClient.createOrigin(
        accessToken: widget.session.accessToken,
        uri: uriController.text.trim(),
        rules: rules.isEmpty ? null : rules,
      );

      if (!mounted) {
        return;
      }
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(message)));
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

  Future<void> _updateRules(OriginItem origin) async {
    final rules = await _openRuleBuilder();
    if (rules == null) {
      return;
    }

    try {
      final message = await widget.apiClient.updateOriginRules(
        accessToken: widget.session.accessToken,
        originId: origin.id,
        rules: rules,
      );
      if (!mounted) {
        return;
      }
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(message)));
    } catch (error) {
      if (!mounted) {
        return;
      }
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(error.toString())));
    }
  }

  Future<void> _triggerScan(OriginItem origin) async {
    try {
      final scan = await widget.apiClient.triggerOriginScan(
        accessToken: widget.session.accessToken,
        originId: origin.id,
      );
      if (!mounted) {
        return;
      }

      setState(() {
        _upsertScan(scan);
      });

      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Scan ${scan.id} started (${scan.status})')),
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

  Future<void> _cancelScan(OriginItem origin) async {
    try {
      final scan = await widget.apiClient.cancelOriginScan(
        accessToken: widget.session.accessToken,
        originId: origin.id,
      );
      if (!mounted) {
        return;
      }

      setState(() {
        _upsertScan(scan);
      });

      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Scan ${scan.id} cancelled (${scan.status})')),
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

  void _toggleExpanded(OriginItem origin) {
    final isExpanded = _expanded[origin.id] ?? false;
    setState(() {
      _expanded[origin.id] = !isExpanded;
    });

    if (!isExpanded) {
      _fetchLatestScan(origin.id);
    }
  }

  String _formatTimestamp(String raw) {
    final parsed = DateTime.tryParse(raw);
    if (parsed == null) {
      return raw;
    }
    final local = parsed.toLocal();
    final month = local.month.toString().padLeft(2, '0');
    final day = local.day.toString().padLeft(2, '0');
    final hour = local.hour.toString().padLeft(2, '0');
    final minute = local.minute.toString().padLeft(2, '0');
    return '${local.year}-$month-$day $hour:$minute';
  }

  ButtonStyle _originActionButtonStyle() {
    return OutlinedButton.styleFrom(
      foregroundColor: Colors.white,
      backgroundColor: Colors.white.withValues(alpha: 0.06),
      side: const BorderSide(color: Colors.white54),
      textStyle: const TextStyle(fontWeight: FontWeight.w600),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Align(
          alignment: Alignment.centerRight,
          child: FilledButton.icon(
            onPressed: _createOrigin,
            icon: const Icon(Icons.add),
            label: const Text('Create origin'),
          ),
        ),
        const SizedBox(height: 12),
        Expanded(
          child: FutureBuilder<List<OriginItem>>(
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

              final origins = snapshot.data ?? const <OriginItem>[];
              if (origins.isEmpty) {
                return const _EmptyPane(title: 'No origins found');
              }

              return RefreshIndicator(
                onRefresh: _refresh,
                child: ListView.separated(
                  itemCount: origins.length,
                  separatorBuilder: (_, _) => const SizedBox(height: 10),
                  itemBuilder: (context, index) {
                    final origin = origins[index];
                    final scans = _scanHistoryByOrigin[origin.id] ?? const [];
                    final expanded = _expanded[origin.id] ?? false;

                    return Card(
                      child: Padding(
                        padding: const EdgeInsets.all(14),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Row(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Container(
                                  width: 42,
                                  height: 42,
                                  decoration: BoxDecoration(
                                    borderRadius: BorderRadius.circular(12),
                                    color: Colors.white.withValues(alpha: 0.08),
                                    border: Border.all(color: Colors.white24),
                                  ),
                                  child: const Icon(Icons.source_outlined),
                                ),
                                const SizedBox(width: 12),
                                Expanded(
                                  child: Column(
                                    crossAxisAlignment:
                                        CrossAxisAlignment.start,
                                    children: [
                                      Text(
                                        origin.uri,
                                        style: const TextStyle(
                                          fontSize: 16,
                                          fontWeight: FontWeight.w700,
                                        ),
                                        maxLines: 2,
                                        overflow: TextOverflow.ellipsis,
                                      ),
                                      const SizedBox(height: 10),
                                      Wrap(
                                        spacing: 14,
                                        runSpacing: 6,
                                        children: [
                                          Text(
                                            'ID ${origin.id}',
                                            style: TextStyle(
                                              color: Colors.white.withValues(
                                                alpha: 0.72,
                                              ),
                                            ),
                                          ),
                                          Text(
                                            'Created ${_formatTimestamp(origin.createdAt)}',
                                            style: TextStyle(
                                              color: Colors.white.withValues(
                                                alpha: 0.72,
                                              ),
                                            ),
                                          ),
                                          Text(
                                            '${scans.length} scans',
                                            style: TextStyle(
                                              color: Colors.white.withValues(
                                                alpha: 0.72,
                                              ),
                                            ),
                                          ),
                                        ],
                                      ),
                                    ],
                                  ),
                                ),
                                const SizedBox(width: 8),
                                IconButton(
                                  tooltip: expanded
                                      ? 'Hide scans'
                                      : 'Show scans',
                                  onPressed: () => _toggleExpanded(origin),
                                  icon: Icon(
                                    expanded
                                        ? Icons.visibility_off_outlined
                                        : Icons.visibility_outlined,
                                  ),
                                ),
                              ],
                            ),
                            const SizedBox(height: 12),
                            Wrap(
                              spacing: 8,
                              runSpacing: 8,
                              children: [
                                OutlinedButton.icon(
                                  onPressed: () => _triggerScan(origin),
                                  style: _originActionButtonStyle(),
                                  icon: const Icon(Icons.play_circle_outline),
                                  label: const Text('Start scan'),
                                ),
                                OutlinedButton.icon(
                                  onPressed: () => _updateRules(origin),
                                  style: _originActionButtonStyle(),
                                  icon: const Icon(Icons.rule),
                                  label: const Text('Rules'),
                                ),
                              ],
                            ),
                            if (expanded) ...[
                              const Divider(),
                              Row(
                                children: [
                                  const Text(
                                    'Scan details',
                                    style: TextStyle(
                                      fontWeight: FontWeight.w700,
                                    ),
                                  ),
                                  const Spacer(),
                                  OutlinedButton.icon(
                                    onPressed: () =>
                                        _fetchLatestScan(origin.id),
                                    style: _originActionButtonStyle(),
                                    icon: const Icon(Icons.refresh),
                                    label: const Text('Refresh'),
                                  ),
                                ],
                              ),
                              const SizedBox(height: 8),
                              if (scans.isEmpty)
                                const Text(
                                  'No scans captured yet for this origin.',
                                )
                              else
                                ListView.separated(
                                  shrinkWrap: true,
                                  physics: const NeverScrollableScrollPhysics(),
                                  itemCount: scans.length,
                                  separatorBuilder: (_, _) =>
                                      const SizedBox(height: 8),
                                  itemBuilder: (context, index) {
                                    final scan = scans[index];
                                    return Container(
                                      width: double.infinity,
                                      padding: const EdgeInsets.all(10),
                                      decoration: BoxDecoration(
                                        borderRadius: BorderRadius.circular(12),
                                        color: Colors.white.withValues(
                                          alpha: 0.04,
                                        ),
                                        border: Border.all(
                                          color: Colors.white24,
                                        ),
                                      ),
                                      child: Row(
                                        children: [
                                          Expanded(
                                            child: Column(
                                              crossAxisAlignment:
                                                  CrossAxisAlignment.start,
                                              children: [
                                                Text(
                                                  'Scan #${scan.id}',
                                                  style: const TextStyle(
                                                    fontWeight: FontWeight.w700,
                                                  ),
                                                ),
                                                const SizedBox(height: 4),
                                                Text(
                                                  'Status: ${scan.status} • Attempts: ${scan.attempts}',
                                                ),
                                                const SizedBox(height: 2),
                                                Text(
                                                  'Discovered files: ${scan.discoveredFiles}',
                                                ),
                                                const SizedBox(height: 2),
                                                Text(
                                                  'Created: ${_formatTimestamp(scan.createdAt)}',
                                                ),
                                              ],
                                            ),
                                          ),
                                          IconButton(
                                            tooltip: 'Cancel this scan',
                                            onPressed: () =>
                                                _cancelScan(origin),
                                            icon: const Icon(
                                              Icons.cancel_outlined,
                                              color: Colors.white,
                                            ),
                                          ),
                                        ],
                                      ),
                                    );
                                  },
                                ),
                            ],
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

class _RulePatternDraft {
  _RulePatternDraft({
    required this.pattern,
    List<_LabelAssignmentDraft>? assignments,
  }) : assignments = assignments ?? [];

  String pattern;
  List<_LabelAssignmentDraft> assignments;
}

class _LabelAssignmentDraft {
  _LabelAssignmentDraft({required this.label, required this.value});

  String label;
  String value;
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
