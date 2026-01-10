import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:wechat_writing_assistant/features/hot_topics/data/datasources/remote/http_fetcher.dart';
import 'package:wechat_writing_assistant/features/hot_topics/domain/entities/topic_source.dart';
import 'package:wechat_writing_assistant/features/hot_topics/presentation/providers/hot_topics_providers.dart';

void main() {
  group('hot_topics providers', () {
    test('wires adapters + repository + usecases', () async {
      final fetcher = _FakeHttpFetcher(
        responses: {
          'https://weibo.com/ajax/side/hotSearch': const FetchResponse(
            statusCode: 200,
            body: '{"data":{"realtime":[{"word":"Weibo1","raw_hot":1}]}}',
            headers: {},
          ),
          'https://www.zhihu.com/api/v3/feed/topstory/hot-lists/total?limit=50&desktop=true': const FetchResponse(
            statusCode: 200,
            body: '{"data":[{"target":{"title":"Zhihu1","url":"https://www.zhihu.com/q/1"}}]}',
            headers: {},
          ),
          'https://top.baidu.com/api/board?platform=wise&tab=realtime': const FetchResponse(
            statusCode: 200,
            body: '{"data":{"cards":[{"content":[{"word":"Baidu1","hotScore":1}]}]}}',
            headers: {},
          ),
          'https://gateway.36kr.com/api/mis/nav/home/nav/rank/hot': const FetchResponse(
            statusCode: 200,
            body: '{"data":{"hotRankList":[{"title":"36kr1","hotValue":1}]}}',
            headers: {},
          ),
        },
      );

      final container = ProviderContainer(
        overrides: [
          httpFetcherProvider.overrideWithValue(fetcher),
        ],
      );
      addTearDown(container.dispose);

      final getHotTopics = container.read(getHotTopicsUseCaseProvider);
      final topics = await getHotTopics();

      expect(topics, hasLength(4));
      expect(
        topics.map((t) => t.source).toSet(),
        equals({TopicSource.weibo, TopicSource.zhihu, TopicSource.baidu, TopicSource.kr36}),
      );
      expect(fetcher.calls['https://weibo.com/ajax/side/hotSearch'], 1);

      final topics2 = await getHotTopics();
      expect(topics2, hasLength(4));
      expect(fetcher.calls['https://weibo.com/ajax/side/hotSearch'], 1, reason: 'should hit repository cache');

      final searchHotTopics = container.read(searchHotTopicsUseCaseProvider);
      final searched = await searchHotTopics('1');
      expect(searched, hasLength(4));
      expect(fetcher.calls['https://weibo.com/ajax/side/hotSearch'], 2, reason: 'search falls back to fetching list');
    });

    test('throws when all sources fail', () async {
      final fetcher = _FakeHttpFetcher(
        responses: {
          'https://weibo.com/ajax/side/hotSearch': const FetchResponse(statusCode: 500, body: 'x', headers: {}),
          'https://www.zhihu.com/api/v3/feed/topstory/hot-lists/total?limit=50&desktop=true':
              const FetchResponse(statusCode: 500, body: 'x', headers: {}),
          'https://top.baidu.com/api/board?platform=wise&tab=realtime':
              const FetchResponse(statusCode: 500, body: 'x', headers: {}),
          'https://gateway.36kr.com/api/mis/nav/home/nav/rank/hot':
              const FetchResponse(statusCode: 500, body: 'x', headers: {}),
        },
      );

      final container = ProviderContainer(
        overrides: [
          httpFetcherProvider.overrideWithValue(fetcher),
        ],
      );
      addTearDown(container.dispose);

      final getHotTopics = container.read(getHotTopicsUseCaseProvider);
      await expectLater(getHotTopics(), throwsA(isA<Exception>()));
    });
  });
}

class _FakeHttpFetcher implements HttpFetcher {
  _FakeHttpFetcher({required this.responses});

  final Map<String, FetchResponse> responses;
  final Map<String, int> calls = {};

  @override
  Future<FetchResponse> get(
    Uri uri, {
    Map<String, String> headers = const {},
    Duration timeout = const Duration(seconds: 15),
  }) async {
    final key = uri.toString();
    calls[key] = (calls[key] ?? 0) + 1;
    return responses[key] ?? const FetchResponse(statusCode: 404, body: '', headers: {});
  }
}
