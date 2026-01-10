import 'package:flutter_test/flutter_test.dart';
import 'package:wechat_writing_assistant/features/hot_topics/domain/entities/hot_topic.dart';
import 'package:wechat_writing_assistant/features/hot_topics/domain/entities/topic_source.dart';
import 'package:wechat_writing_assistant/features/hot_topics/presentation/providers/hot_topics_state.dart';

void main() {
  test('HotTopicsViewState copyWith updates fields', () {
    final state = HotTopicsViewState(
      topics: [
        HotTopic(
          id: 'a',
          source: TopicSource.weibo,
          rank: 1,
          title: 'A',
          fetchedAt: DateTime(2026, 1, 1),
        ),
      ],
      selectedSource: null,
      query: '',
      updatedAt: DateTime(2026, 1, 1),
    );

    final next = state.copyWith(query: 'flutter', selectedSource: TopicSource.zhihu);
    expect(next.query, 'flutter');
    expect(next.selectedSource, TopicSource.zhihu);
    expect(next.topics, state.topics);
  });
}
