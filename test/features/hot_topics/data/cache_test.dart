import 'package:flutter_test/flutter_test.dart';
import 'package:wechat_writing_assistant/features/hot_topics/data/datasources/cache/hot_topics_cache.dart';
import 'package:wechat_writing_assistant/features/hot_topics/data/datasources/cache/ttl_cache.dart';
import 'package:wechat_writing_assistant/features/hot_topics/data/models/hot_topic_model.dart';
import 'package:wechat_writing_assistant/features/hot_topics/domain/entities/topic_source.dart';

void main() {
  group('TTL cache', () {
    test('TtlCache expires after ttl', () {
      var now = DateTime(2026, 1, 1, 0, 0, 0);
      final cache = TtlCache<String, int>(
        ttl: const Duration(seconds: 10),
        now: () => now,
      );

      cache.write('a', 1);
      expect(cache.read('a'), 1);

      now = now.add(const Duration(seconds: 11));
      expect(cache.read('a'), isNull);
      expect(cache.size, 0);
    });

    test('HotTopicsCache caches per source and query', () {
      var now = DateTime(2026, 1, 1, 0, 0, 0);
      final cache = HotTopicsCache(
        ttl: const Duration(seconds: 10),
        now: () => now,
      );

      final topics = [
        HotTopicModel.fromParts(
          source: TopicSource.weibo,
          rank: 1,
          title: 'A',
          fetchedAt: now,
        ),
      ];

      cache.writeHotTopics(source: TopicSource.weibo, topics: topics);
      expect(cache.readHotTopics(source: TopicSource.weibo), isNotNull);
      expect(cache.readHotTopics(source: TopicSource.zhihu), isNull);

      cache.writeSearch(source: TopicSource.weibo, query: 'Flutter', topics: topics);
      expect(cache.readSearch(source: TopicSource.weibo, query: 'flutter'), isNotNull);
      expect(cache.readSearch(source: TopicSource.weibo, query: 'other'), isNull);

      now = now.add(const Duration(seconds: 11));
      expect(cache.readHotTopics(source: TopicSource.weibo), isNull);
      expect(cache.readSearch(source: TopicSource.weibo, query: 'flutter'), isNull);
    });
  });
}
