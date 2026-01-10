import 'package:flutter_test/flutter_test.dart';
import 'package:wechat_writing_assistant/features/hot_topics/data/datasources/cache/hot_topics_cache.dart';
import 'package:wechat_writing_assistant/features/hot_topics/data/datasources/hot_topic_source.dart';
import 'package:wechat_writing_assistant/features/hot_topics/data/models/hot_topic_model.dart';
import 'package:wechat_writing_assistant/features/hot_topics/data/repositories/hot_topics_repository_impl.dart';
import 'package:wechat_writing_assistant/features/hot_topics/domain/entities/topic_source.dart';

void main() {
  group('HotTopicsRepositoryImpl', () {
    test('caches per-source hot topics and sorts by rank', () async {
      final source = _FakeSource(
        source: TopicSource.weibo,
        hotTopics: [
          _topic(TopicSource.weibo, 2, 'B'),
          _topic(TopicSource.weibo, 1, 'A'),
        ],
      );
      final cache = HotTopicsCache(ttl: const Duration(minutes: 10));
      final repo = HotTopicsRepositoryImpl(sources: [source], cache: cache);

      final first = await repo.getHotTopics(source: TopicSource.weibo);
      expect(source.fetchCount, 1);
      expect(first.map((t) => t.rank).toList(), [1, 2]);

      final second = await repo.getHotTopics(source: TopicSource.weibo);
      expect(source.fetchCount, 1, reason: 'should hit cache');
      expect(second.map((t) => t.title).toList(), ['A', 'B']);

      final third = await repo.getHotTopics(source: TopicSource.weibo, forceRefresh: true);
      expect(source.fetchCount, 2);
      expect(third, isNotEmpty);
    });

    test('merges sources in configured order', () async {
      final zhihu = _FakeSource(
        source: TopicSource.zhihu,
        hotTopics: [_topic(TopicSource.zhihu, 1, 'Z1')],
      );
      final weibo = _FakeSource(
        source: TopicSource.weibo,
        hotTopics: [_topic(TopicSource.weibo, 1, 'W1')],
      );

      final repo = HotTopicsRepositoryImpl(
        sources: [zhihu, weibo],
        cache: HotTopicsCache(ttl: const Duration(minutes: 10)),
      );

      final topics = await repo.getHotTopics();
      expect(topics.map((t) => t.source).toList(), [TopicSource.zhihu, TopicSource.weibo]);
    });

    test('refreshHotTopics invalidates cache', () async {
      final source = _FakeSource(
        source: TopicSource.baidu,
        hotTopics: [_topic(TopicSource.baidu, 1, 'A')],
      );
      final repo = HotTopicsRepositoryImpl(
        sources: [source],
        cache: HotTopicsCache(ttl: const Duration(minutes: 10)),
      );

      await repo.getHotTopics(source: TopicSource.baidu);
      expect(source.fetchCount, 1);

      await repo.refreshHotTopics(source: TopicSource.baidu);
      expect(source.fetchCount, 2);
    });

    test('searchHotTopics caches by normalized query', () async {
      final source = _FakeSource(
        source: TopicSource.kr36,
        hotTopics: [
          _topic(TopicSource.kr36, 1, 'Flutter 3.10'),
          _topic(TopicSource.kr36, 2, 'Dart'),
        ],
      );
      final repo = HotTopicsRepositoryImpl(
        sources: [source],
        cache: HotTopicsCache(ttl: const Duration(minutes: 10)),
      );

      final first = await repo.searchHotTopics('flutter', source: TopicSource.kr36);
      expect(source.searchCount, 1);
      expect(first, hasLength(1));

      final second = await repo.searchHotTopics('  Flutter  ', source: TopicSource.kr36);
      expect(source.searchCount, 1, reason: 'should hit cache');
      expect(second, hasLength(1));
    });
  });
}

HotTopicModel _topic(TopicSource source, int rank, String title) {
  return HotTopicModel.fromParts(
    source: source,
    rank: rank,
    title: title,
    fetchedAt: DateTime(2026, 1, 1),
  );
}

class _FakeSource implements HotTopicSource {
  _FakeSource({
    required this.source,
    required List<HotTopicModel> hotTopics,
  }) : _hotTopics = hotTopics;

  @override
  final TopicSource source;

  final List<HotTopicModel> _hotTopics;

  int fetchCount = 0;
  int searchCount = 0;

  @override
  Future<List<HotTopicModel>> fetchHotTopics() async {
    fetchCount++;
    return _hotTopics;
  }

  @override
  Future<List<HotTopicModel>> search(String query) async {
    searchCount++;
    final normalized = query.trim().toLowerCase();
    return _hotTopics
        .where((t) => t.title.toLowerCase().contains(normalized))
        .toList(growable: false);
  }
}
