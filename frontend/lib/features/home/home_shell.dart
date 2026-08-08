import 'dart:convert';

import 'package:flutter/material.dart';

import '../../colors.dart';
import '../../models.dart';
import '../../services/api_client.dart';
import '../views/blob_history_view.dart';
import '../views/blobs_view.dart';
import '../views/datasets_view.dart';
import '../views/labels_view.dart';
import '../views/origins_view.dart';
import '../views/settings_view.dart';
import '../views/users_view.dart';

class HomeShell extends StatefulWidget {
  const HomeShell({
    super.key,
    required this.apiClient,
    required this.session,
    required this.currentApiUrl,
    required this.onApiUrlChanged,
    required this.onSessionChanged,
    required this.onLogout,
  });

  final ApiClient apiClient;
  final AuthSession session;
  final String currentApiUrl;
  final ValueChanged<String> onApiUrlChanged;
  final ValueChanged<AuthSession> onSessionChanged;
  final VoidCallback onLogout;

  @override
  State<HomeShell> createState() => _HomeShellState();
}

class _HomeShellState extends State<HomeShell> {
  int _selectedIndex = 0;
  bool _canManageUsers = false;
  bool _isNavVisible = true;

  @override
  void initState() {
    super.initState();
    _loadPermissions();
  }

  @override
  void didUpdateWidget(covariant HomeShell oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.session.accessToken != widget.session.accessToken) {
      _loadPermissions();
    }
  }

  Future<void> _loadPermissions() async {
    final currentUserId = _extractUserIdFromToken(widget.session.accessToken);
    if (currentUserId == null) {
      if (mounted) {
        setState(() {
          _canManageUsers = false;
          _selectedIndex = 0;
        });
      }
      return;
    }

    try {
      final users = await widget.apiClient.getUsers(widget.session.accessToken);
      final currentUser = users
          .where((user) => user.id == currentUserId)
          .cast<UserItem?>()
          .firstWhere((user) => user != null, orElse: () => null);

      final canManage = currentUser?.role.toLowerCase() == 'admin';
      if (!mounted) {
        return;
      }

      setState(() {
        _canManageUsers = canManage;
        final sections = _sections();
        if (_selectedIndex >= sections.length) {
          _selectedIndex = 0;
        }
      });
    } catch (_) {
      if (!mounted) {
        return;
      }
      setState(() {
        _canManageUsers = false;
        final sections = _sections();
        if (_selectedIndex >= sections.length) {
          _selectedIndex = 0;
        }
      });
    }
  }

  int? _extractUserIdFromToken(String token) {
    final parts = token.split('.');
    if (parts.length != 3) {
      return null;
    }

    try {
      final payload = utf8.decode(
        base64Url.decode(base64Url.normalize(parts[1])),
      );
      final json = jsonDecode(payload);
      if (json is! Map<String, dynamic>) {
        return null;
      }

      final subject = json['sub'];
      if (subject is String) {
        return int.tryParse(subject);
      }
      if (subject is num) {
        return subject.toInt();
      }
      return null;
    } catch (_) {
      return null;
    }
  }

  List<_NavSection> _sections() {
    final sections = <_NavSection>[
      _NavSection(
        title: 'Blobs',
        destination: const NavigationRailDestination(
          icon: Icon(Icons.search_outlined),
          selectedIcon: Icon(Icons.search),
          label: Text('Blobs'),
        ),
        page: BlobsView(apiClient: widget.apiClient, session: widget.session),
      ),
      _NavSection(
        title: 'History',
        destination: const NavigationRailDestination(
          icon: Icon(Icons.history_outlined),
          selectedIcon: Icon(Icons.history),
          label: Text('History'),
        ),
        page: BlobHistoryView(
          apiClient: widget.apiClient,
          session: widget.session,
        ),
      ),
      _NavSection(
        title: 'Datasets',
        destination: const NavigationRailDestination(
          icon: Icon(Icons.inventory_2_outlined),
          selectedIcon: Icon(Icons.inventory_2),
          label: Text('Datasets'),
        ),
        page: DatasetsView(
          apiClient: widget.apiClient,
          session: widget.session,
        ),
      ),
      _NavSection(
        title: 'Origins',
        destination: const NavigationRailDestination(
          icon: Icon(Icons.source_outlined),
          selectedIcon: Icon(Icons.source),
          label: Text('Origins'),
        ),
        page: OriginsView(apiClient: widget.apiClient, session: widget.session),
      ),
      _NavSection(
        title: 'Labels',
        destination: const NavigationRailDestination(
          icon: Icon(Icons.label_outline),
          selectedIcon: Icon(Icons.label),
          label: Text('Labels'),
        ),
        page: LabelsView(apiClient: widget.apiClient, session: widget.session),
      ),
    ];

    if (_canManageUsers) {
      sections.add(
        _NavSection(
          title: 'Users',
          destination: const NavigationRailDestination(
            icon: Icon(Icons.group_outlined),
            selectedIcon: Icon(Icons.group),
            label: Text('Users'),
          ),
          page: UsersView(apiClient: widget.apiClient, session: widget.session),
        ),
      );
    }

    sections.add(
      _NavSection(
        title: 'Settings',
        destination: const NavigationRailDestination(
          icon: Icon(Icons.settings_outlined),
          selectedIcon: Icon(Icons.settings),
          label: Text('Settings'),
        ),
        page: SettingsView(
          apiClient: widget.apiClient,
          session: widget.session,
          currentApiUrl: widget.currentApiUrl,
          onApiUrlChanged: widget.onApiUrlChanged,
          onSessionChanged: widget.onSessionChanged,
          onLogout: widget.onLogout,
        ),
      ),
    );

    return sections;
  }

  Widget _buildNavToggleButton() {
    final isVisible = _isNavVisible;

    return Tooltip(
      message: isVisible ? 'Hide navigation' : 'Show navigation',
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: () {
            setState(() {
              _isNavVisible = !isVisible;
            });
          },
          borderRadius: BorderRadius.circular(isVisible ? 12 : 999),
          child: Ink(
            width: 42,
            height: 42,
            decoration: BoxDecoration(
              color: OdessaColors.primary.withValues(
                alpha: isVisible ? 0.72 : 0.95,
              ),
              borderRadius: BorderRadius.circular(isVisible ? 12 : 999),
              border: Border.all(
                color: OdessaColors.accentDark.withValues(alpha: 0.45),
              ),
            ),
            child: Icon(
              isVisible ? Icons.chevron_left_rounded : Icons.menu_rounded,
              color: Colors.white,
            ),
          ),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final sections = _sections();
    final settingsIndex = sections.indexWhere(
      (section) => section.title == 'Settings',
    );
    final railEntries = _buildRailEntries(
      sections,
      excludedSectionIndex: settingsIndex,
    );
    if (_selectedIndex >= sections.length) {
      _selectedIndex = 0;
    }

    final selectedRailIndex = railEntries.indexWhere(
      (entry) => entry.sectionIndex == _selectedIndex,
    );

    return Scaffold(
      body: Row(
        children: [
          Container(
            width: _isNavVisible ? 88 : 56,
            color: OdessaColors.background,
            child: Column(
              children: [
                const SizedBox(height: 8),
                if (_isNavVisible)
                  Expanded(
                    child: NavigationRail(
                      selectedIndex: selectedRailIndex < 0
                          ? 0
                          : selectedRailIndex,
                      useIndicator: true,
                      onDestinationSelected: (index) {
                        final entry = railEntries[index];
                        if (entry.sectionIndex == null) {
                          return;
                        }
                        setState(() {
                          _selectedIndex = entry.sectionIndex!;
                        });
                      },
                      labelType: NavigationRailLabelType.all,
                      indicatorColor: OdessaColors.accentDark.withValues(
                        alpha: 0.24,
                      ),
                      selectedIconTheme: const IconThemeData(
                        color: Colors.white,
                      ),
                      selectedLabelTextStyle: const TextStyle(
                        color: Colors.white,
                        fontWeight: FontWeight.w700,
                      ),
                      leading: const SizedBox(height: 12),
                      destinations: railEntries
                          .map((entry) => entry.destination)
                          .toList(),
                    ),
                  )
                else
                  const Spacer(),
                if (settingsIndex >= 0)
                  Padding(
                    padding: const EdgeInsets.only(bottom: 8),
                    child: Tooltip(
                      message: 'Settings',
                      child: IconButton(
                        onPressed: () {
                          setState(() {
                            _selectedIndex = settingsIndex;
                          });
                        },
                        icon: Icon(
                          _selectedIndex == settingsIndex
                              ? Icons.settings
                              : Icons.settings_outlined,
                        ),
                      ),
                    ),
                  ),
                Padding(
                  padding: const EdgeInsets.only(bottom: 8),
                  child: Tooltip(
                    message: 'Logout',
                    child: IconButton(
                      onPressed: widget.onLogout,
                      icon: const Icon(Icons.logout),
                    ),
                  ),
                ),
                Padding(
                  padding: const EdgeInsets.only(bottom: 12),
                  child: _buildNavToggleButton(),
                ),
              ],
            ),
          ),
          const VerticalDivider(width: 1),
          Expanded(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Padding(
                    padding: const EdgeInsets.symmetric(vertical: 8),
                    child: Text(
                      sections[_selectedIndex].title,
                      textAlign: TextAlign.center,
                      style: Theme.of(context).textTheme.headlineSmall
                          ?.copyWith(
                            color: OdessaColors.accentDark,
                            fontWeight: FontWeight.w700,
                          ),
                    ),
                  ),
                  const SizedBox(height: 12),
                  Expanded(child: sections[_selectedIndex].page),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

List<_RailEntry> _buildRailEntries(
  List<_NavSection> sections, {
  required int excludedSectionIndex,
}) {
  final entries = <_RailEntry>[];
  for (var i = 0; i < sections.length; i++) {
    if (i == excludedSectionIndex) {
      continue;
    }
    if (sections[i].title == 'Datasets') {
      entries.add(_RailEntry.spacer());
    }
    entries.add(
      _RailEntry(sectionIndex: i, destination: sections[i].destination),
    );
  }
  return entries;
}

class _NavSection {
  const _NavSection({
    required this.title,
    required this.destination,
    required this.page,
  });

  final String title;
  final NavigationRailDestination destination;
  final Widget page;
}

class _RailEntry {
  const _RailEntry({required this.sectionIndex, required this.destination});

  const _RailEntry.spacer()
    : sectionIndex = null,
      destination = const NavigationRailDestination(
        icon: SizedBox(height: 10),
        selectedIcon: SizedBox(height: 10),
        label: Text(''),
      );

  final int? sectionIndex;
  final NavigationRailDestination destination;
}
