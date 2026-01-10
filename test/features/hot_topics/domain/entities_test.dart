import 'package:flutter_test/flutter_test.dart';
import 'package:wechat_writing_assistant/features/hot_topics/domain/entities/hot_topic.dart';
import 'package:wechat_writing_assistant/features/hot_topics/domain/entities/topic_source.dart';

void main() {
  group('Domain entities', () {
    test('TopicSource provides label/key/homepage', () {
      for (final source in TopicSource.values) {
        expect(source.key, isNotEmpty);
        expect(source.label, isNotEmpty);
        expect(source.homepage, isNotNull);
      }
    });

    test('HotTopic supports copyWith, equality and hashCode', () {
      final a = HotTopic(
        id: 'id1',
        source: TopicSource.weibo,
        rank: 1,
        title: 'A',
        url: Uri.parse('https://example.com/a'),
        hotValue: 123,
        description: 'desc',
        fetchedAt: DateTime(2026, 1, 1),
      );

      final b = a.copyWith();
      expect(b, equals(a));
      expect(b.hashCode, equals(a.hashCode));

      final c = a.copyWith(title: 'B');
      expect(c, isNot(equals(a)));
      expect(c.title, 'B');

      expect(a.toString(), contains('HotTopic'));
      expect(a.toString(), contains('weibo'));
    });
  });
}
