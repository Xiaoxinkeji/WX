import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:wechat_writing_assistant/features/hot_topics/domain/entities/hot_topic.dart';
import 'package:wechat_writing_assistant/features/hot_topics/domain/entities/topic_source.dart';
import 'package:wechat_writing_assistant/features/hot_topics/domain/repositories/hot_topics_repository.dart';
import 'package:wechat_writing_assistant/features/hot_topics/presentation/pages/hot_topics_page.dart';
import 'package:wechat_writing_assistant/features/hot_topics/presentation/providers/hot_topics_providers.dart';

void main() {
  testWidgets('HotTopicsPage shows error state', (tester) async {
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          hotTopicsRepositoryProvider.overrideWithValue(_ThrowingRepo()),
        ],
        child: const MaterialApp(home: HotTopicsPage()),
      ),
    );

    await tester.pump();
    await tester.pump();

    expect(find.text('加载失败'), findsOneWidget);
    expect(find.text('重试'), findsOneWidget);
  });
}

class _ThrowingRepo implements HotTopicsRepository {
  @override
  Future<List<HotTopic>> getHotTopics({TopicSource? source, bool forceRefresh = false}) async {
    throw Exception('boom');
  }

  @override
  Future<List<HotTopic>> refreshHotTopics({TopicSource? source}) async {
    throw Exception('boom');
  }

  @override
  Future<List<HotTopic>> searchHotTopics(String query,
      {TopicSource? source, bool forceRefresh = false}) async {
    throw Exception('boom');
  }
}
