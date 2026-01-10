import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:wechat_writing_assistant/features/hot_topics/domain/entities/hot_topic.dart';
import 'package:wechat_writing_assistant/features/hot_topics/domain/entities/topic_source.dart';
import 'package:wechat_writing_assistant/features/hot_topics/domain/repositories/hot_topics_repository.dart';
import 'package:wechat_writing_assistant/features/hot_topics/presentation/providers/hot_topics_providers.dart';

void main() {
  group('HotTopicsController', () {
    test('build loads initial topics', () async {
      final repo = _FakeRepo();
      final container = ProviderContainer(
        overrides: [
          hotTopicsRepositoryProvider.overrideWithValue(repo),
        ],
      );
      addTearDown(container.dispose);

      final state = await container.read(hotTopicsControllerProvider.future);
      expect(state.query, '');
      expect(state.selectedSource, isNull);
      expect(state.topics, hasLength(2));
      expect(repo.getCalls, 1);
    });

    test('setSource triggers source-specific fetch', () async {
      final repo = _FakeRepo();
      final container = ProviderContainer(
        overrides: [
          hotTopicsRepositoryProvider.overrideWithValue(repo),
        ],
      );
      addTearDown(container.dispose);

      await container.read(hotTopicsControllerProvider.future);
      await container.read(hotTopicsControllerProvider.notifier).setSource(TopicSource.weibo);

      final current = container.read(hotTopicsControllerProvider).valueOrNull;
      expect(current, isNotNull);
      expect(current!.selectedSource, TopicSource.weibo);
      expect(repo.lastGetSource, TopicSource.weibo);
    });

    test('search stores query and calls repository search', () async {
      final repo = _FakeRepo();
      final container = ProviderContainer(
        overrides: [
          hotTopicsRepositoryProvider.overrideWithValue(repo),
        ],
      );
      addTearDown(container.dispose);

      await container.read(hotTopicsControllerProvider.future);
      await container.read(hotTopicsControllerProvider.notifier).search('flutter');

      final current = container.read(hotTopicsControllerProvider).valueOrNull;
      expect(current, isNotNull);
      expect(current!.query, 'flutter');
      expect(repo.searchCalls, 1);
    });

    test('refresh uses search when query is non-empty', () async {
      final repo = _FakeRepo();
      final container = ProviderContainer(
        overrides: [
          hotTopicsRepositoryProvider.overrideWithValue(repo),
        ],
      );
      addTearDown(container.dispose);

      await container.read(hotTopicsControllerProvider.future);
      await container.read(hotTopicsControllerProvider.notifier).search('flutter');
      await container.read(hotTopicsControllerProvider.notifier).refresh();

      expect(repo.searchCalls, greaterThanOrEqualTo(2));
      expect(repo.refreshCalls, 0);
    });

    test('refresh uses refreshHotTopics when query is empty', () async {
      final repo = _FakeRepo();
      final container = ProviderContainer(
        overrides: [
          hotTopicsRepositoryProvider.overrideWithValue(repo),
        ],
      );
      addTearDown(container.dispose);

      await container.read(hotTopicsControllerProvider.future);
      await container.read(hotTopicsControllerProvider.notifier).refresh();

      expect(repo.refreshCalls, 1);
    });
  });
}

class _FakeRepo implements HotTopicsRepository {
  int getCalls = 0;
  int refreshCalls = 0;
  int searchCalls = 0;
  TopicSource? lastGetSource;

  @override
  Future<List<HotTopic>> getHotTopics({TopicSource? source, bool forceRefresh = false}) async {
    getCalls++;
    lastGetSource = source;
    return [
      HotTopic(
        id: 'a',
        source: source ?? TopicSource.weibo,
        rank: 1,
        title: 'Flutter',
        url: Uri.parse('https://example.com/a'),
        fetchedAt: DateTime(2026, 1, 1),
      ),
      HotTopic(
        id: 'b',
        source: source ?? TopicSource.zhihu,
        rank: 2,
        title: 'Dart',
        url: Uri.parse('https://example.com/b'),
        fetchedAt: DateTime(2026, 1, 1),
      ),
    ];
  }

  @override
  Future<List<HotTopic>> refreshHotTopics({TopicSource? source}) async {
    refreshCalls++;
    return getHotTopics(source: source, forceRefresh: true);
  }

  @override
  Future<List<HotTopic>> searchHotTopics(String query,
      {TopicSource? source, bool forceRefresh = false}) async {
    searchCalls++;
    final topics = await getHotTopics(source: source, forceRefresh: forceRefresh);
    final normalized = query.toLowerCase();
    return topics.where((t) => t.title.toLowerCase().contains(normalized)).toList();
  }
}
