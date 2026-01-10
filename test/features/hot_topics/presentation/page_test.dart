import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:wechat_writing_assistant/features/hot_topics/domain/entities/hot_topic.dart';
import 'package:wechat_writing_assistant/features/hot_topics/domain/entities/topic_source.dart';
import 'package:wechat_writing_assistant/features/hot_topics/domain/repositories/hot_topics_repository.dart';
import 'package:wechat_writing_assistant/features/hot_topics/presentation/pages/hot_topics_page.dart';
import 'package:wechat_writing_assistant/features/hot_topics/presentation/providers/hot_topics_providers.dart';

void main() {
  testWidgets('HotTopicsPage renders list, filters by search and source', (tester) async {
    final repo = _PageFakeRepo();

    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          hotTopicsRepositoryProvider.overrideWithValue(repo),
        ],
        child: const MaterialApp(home: HotTopicsPage()),
      ),
    );

    // 1st pump: schedules provider load. 2nd pump: resolves async microtasks.
    await tester.pump();
    await tester.pump();

    expect(find.text('Flutter 热点'), findsOneWidget);
    expect(find.text('Dart 热点'), findsOneWidget);

    await tester.enterText(find.byKey(const Key('hotTopicsSearchField')), 'Flutter');
    await tester.testTextInput.receiveAction(TextInputAction.search);
    await tester.pump();
    await tester.pump();

    expect(find.text('Flutter 热点'), findsOneWidget);
    expect(find.text('Dart 热点'), findsNothing);

    final dropdown = find.descendant(
      of: find.byKey(const Key('hotTopicsSourceDropdown')),
      matching: find.byWidgetPredicate((w) => w is DropdownButton),
    );
    await tester.tap(dropdown);
    await tester.pumpAndSettle();

    await tester.tap(find.text('微博').last);
    await tester.pump();
    await tester.pump();

    expect(find.text('Flutter 热点'), findsOneWidget);
    expect(repo.lastGetSource, TopicSource.weibo);
  });
}

class _PageFakeRepo implements HotTopicsRepository {
  TopicSource? lastGetSource;

  @override
  Future<List<HotTopic>> getHotTopics({TopicSource? source, bool forceRefresh = false}) async {
    lastGetSource = source;
    if (source == TopicSource.weibo) {
      return [
        _topic('w1', TopicSource.weibo, 1, 'Flutter 热点'),
        _topic('w2', TopicSource.weibo, 2, 'Dart 热点'),
      ];
    }

    return [
      _topic('z1', TopicSource.zhihu, 1, 'Flutter 热点'),
      _topic('b1', TopicSource.baidu, 1, 'Dart 热点'),
    ];
  }

  @override
  Future<List<HotTopic>> refreshHotTopics({TopicSource? source}) async {
    return getHotTopics(source: source, forceRefresh: true);
  }

  @override
  Future<List<HotTopic>> searchHotTopics(
    String query, {
    TopicSource? source,
    bool forceRefresh = false,
  }) async {
    final normalized = query.trim().toLowerCase();
    final topics = await getHotTopics(source: source, forceRefresh: forceRefresh);
    return topics.where((t) => t.title.toLowerCase().contains(normalized)).toList(growable: false);
  }

  HotTopic _topic(String id, TopicSource source, int rank, String title) {
    return HotTopic(
      id: id,
      source: source,
      rank: rank,
      title: title,
      url: Uri.parse('https://example.com/$id'),
      fetchedAt: DateTime(2026, 1, 1),
    );
  }
}

