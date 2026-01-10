import 'package:flutter_test/flutter_test.dart';
import 'package:wechat_writing_assistant/features/hot_topics/data/models/hot_topic_model.dart';
import 'package:wechat_writing_assistant/features/hot_topics/domain/entities/topic_source.dart';

void main() {
  group('HotTopicModel', () {
    test('generateId uses url when available', () {
      final id = HotTopicModel.generateId(
        source: TopicSource.zhihu,
        title: ' Hello  World ',
        url: Uri.parse('https://example.com/x'),
      );
      expect(id, startsWith('${TopicSource.zhihu.key}:'));
      expect(id, contains('https://example.com/x'));
    });

    test('generateId falls back to normalized title', () {
      final id = HotTopicModel.generateId(
        source: TopicSource.baidu,
        title: ' Hello   World ',
      );
      expect(id, equals('${TopicSource.baidu.key}:hello world'));
    });

    test('toMap/fromMap round-trip', () {
      final now = DateTime(2026, 1, 1);
      final model = HotTopicModel.fromParts(
        source: TopicSource.kr36,
        rank: 3,
        title: 'Title',
        url: Uri.parse('https://36kr.com/p/1'),
        hotValue: 999,
        description: 'desc',
        fetchedAt: now,
      );

      final map = model.toMap();
      final restored = HotTopicModel.fromMap(Map<String, Object?>.from(map));

      expect(restored.id, model.id);
      expect(restored.source, model.source);
      expect(restored.rank, model.rank);
      expect(restored.title, model.title);
      expect(restored.url, model.url);
      expect(restored.hotValue, model.hotValue);
      expect(restored.description, model.description);
      expect(restored.fetchedAt, model.fetchedAt);
    });

    test('fromMap tolerates unknown source key', () {
      final map = {
        'id': 'x',
        'source': 'unknown',
        'rank': 1,
        'title': 'T',
        'url': null,
        'hotValue': null,
        'description': null,
        'fetchedAt': DateTime(2026, 1, 1).toIso8601String(),
      };

      final restored = HotTopicModel.fromMap(map);
      expect(restored.source, TopicSource.weibo, reason: 'fallback');
    });
  });
}
