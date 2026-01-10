import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/hot_topics_providers.dart';
import '../widgets/hot_topic_card.dart';
import '../widgets/topic_source_dropdown.dart';

class HotTopicsPage extends ConsumerStatefulWidget {
  const HotTopicsPage({super.key});

  @override
  ConsumerState<HotTopicsPage> createState() => _HotTopicsPageState();
}

class _HotTopicsPageState extends ConsumerState<HotTopicsPage> {
  final _searchController = TextEditingController();

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final asyncState = ref.watch(hotTopicsControllerProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('热点扫描'),
        actions: [
          IconButton(
            tooltip: '刷新',
            onPressed: () => ref.read(hotTopicsControllerProvider.notifier).refresh(),
            icon: const Icon(Icons.refresh),
          ),
        ],
      ),
      body: asyncState.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (error, _) {
          return Center(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const Text('加载失败'),
                  const SizedBox(height: 8),
                  Text(
                    '$error',
                    textAlign: TextAlign.center,
                    style: Theme.of(context).textTheme.bodySmall,
                  ),
                  const SizedBox(height: 12),
                  FilledButton(
                    onPressed: () => ref.read(hotTopicsControllerProvider.notifier).refresh(),
                    child: const Text('重试'),
                  ),
                ],
              ),
            ),
          );
        },
        data: (state) {
          if (_searchController.text != state.query) {
            _searchController.text = state.query;
          }

          return Column(
            children: [
              Padding(
                padding: const EdgeInsets.fromLTRB(12, 12, 12, 0),
                child: TextField(
                  key: const Key('hotTopicsSearchField'),
                  controller: _searchController,
                  textInputAction: TextInputAction.search,
                  decoration: InputDecoration(
                    hintText: '搜索热点…',
                    prefixIcon: const Icon(Icons.search),
                    suffixIcon: state.query.isEmpty
                        ? null
                        : IconButton(
                            tooltip: '清空',
                            onPressed: () {
                              _searchController.clear();
                              ref.read(hotTopicsControllerProvider.notifier).search('');
                            },
                            icon: const Icon(Icons.clear),
                          ),
                    border: const OutlineInputBorder(),
                  ),
                  onSubmitted: (q) => ref.read(hotTopicsControllerProvider.notifier).search(q),
                ),
              ),
              Padding(
                padding: const EdgeInsets.fromLTRB(12, 8, 12, 8),
                child: TopicSourceDropdown(
                  key: const Key('hotTopicsSourceDropdown'),
                  value: state.selectedSource,
                  onChanged: (s) => ref.read(hotTopicsControllerProvider.notifier).setSource(s),
                ),
              ),
              Expanded(
                child: RefreshIndicator(
                  onRefresh: () => ref.read(hotTopicsControllerProvider.notifier).refresh(),
                  child: state.topics.isEmpty
                      ? ListView(
                          physics: const AlwaysScrollableScrollPhysics(),
                          children: const [
                            SizedBox(height: 80),
                            Center(child: Text('暂无数据')),
                          ],
                        )
                      : ListView.separated(
                          physics: const AlwaysScrollableScrollPhysics(),
                          itemCount: state.topics.length,
                          separatorBuilder: (_, __) => const SizedBox(height: 4),
                          itemBuilder: (context, index) {
                            return HotTopicCard(topic: state.topics[index]);
                          },
                        ),
                ),
              ),
            ],
          );
        },
      ),
    );
  }
}
