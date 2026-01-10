import 'package:flutter_test/flutter_test.dart';
import 'package:wechat_writing_assistant/features/hot_topics/data/datasources/remote/baidu_hot_topic_source.dart';
import 'package:wechat_writing_assistant/features/hot_topics/data/datasources/remote/kr36_hot_topic_source.dart';
import 'package:wechat_writing_assistant/features/hot_topics/data/datasources/remote/weibo_hot_topic_source.dart';
import 'package:wechat_writing_assistant/features/hot_topics/data/datasources/remote/zhihu_hot_topic_source.dart';
import 'package:wechat_writing_assistant/features/hot_topics/domain/entities/topic_source.dart';

void main() {
  group('Hot topic parsing', () {
    test('Weibo parses realtime list', () {
      const json = r'''
{
  "ok": 1,
  "data": {
    "realtime": [
      {"word": "热点A", "raw_hot": 123, "link": "https://s.weibo.com/weibo?q=A", "rank": 1},
      {"note": "热点B", "raw_hot": "456", "link": "https://s.weibo.com/weibo?q=B"}
    ]
  }
}
''';

      final topics = WeiboHotTopicSource.parseHotTopics(
        json,
        fetchedAt: DateTime(2026, 1, 1),
      );

      expect(topics, hasLength(2));
      expect(topics.first.source, TopicSource.weibo);
      expect(topics.first.rank, 1);
      expect(topics.first.title, '热点A');
      expect(topics.first.url.toString(), contains('weibo?q=A'));
      expect(topics.first.hotValue, 123);

      expect(topics[1].rank, 2);
      expect(topics[1].title, '热点B');
      expect(topics[1].hotValue, 456);
    });

    test('Zhihu parses data list', () {
      const json = r'''
{
  "data": [
    {
      "target": {
        "title": "知乎热榜1",
        "url": "https://www.zhihu.com/question/1",
        "excerpt": "摘要"
      },
      "detail_text": "123 万热度"
    },
    {
      "target": {
        "title": "知乎热榜2",
        "url": "https://www.zhihu.com/question/2"
      }
    }
  ]
}
''';

      final topics = ZhihuHotTopicSource.parseHotTopics(
        json,
        fetchedAt: DateTime(2026, 1, 1),
      );

      expect(topics, hasLength(2));
      expect(topics.first.source, TopicSource.zhihu);
      expect(topics.first.title, '知乎热榜1');
      expect(topics.first.rank, 1);
      expect(topics.first.hotValue, 123);
      expect(topics.first.description, '摘要');
      expect(topics.first.url.toString(), contains('question/1'));

      expect(topics[1].rank, 2);
      expect(topics[1].hotValue, isNull);
    });

    test('Baidu parses board content', () {
      const json = r'''
{
  "data": {
    "cards": [
      {
        "content": [
          {"word": "百度热1", "url": "https://www.baidu.com/s?wd=1", "hotScore": 999, "desc": "描述"},
          {"keyword": "百度热2", "link": "https://www.baidu.com/s?wd=2", "hot_score": "888"}
        ]
      }
    ]
  }
}
''';

      final topics = BaiduHotTopicSource.parseHotTopics(
        json,
        fetchedAt: DateTime(2026, 1, 1),
      );

      expect(topics, hasLength(2));
      expect(topics.first.source, TopicSource.baidu);
      expect(topics.first.title, '百度热1');
      expect(topics.first.hotValue, 999);
      expect(topics.first.description, '描述');

      expect(topics[1].title, '百度热2');
      expect(topics[1].hotValue, 888);
    });

    test('36Kr parses hotRankList', () {
      const json = r'''
{
  "data": {
    "hotRankList": [
      {"title": "36氪热1", "url": "https://36kr.com/p/1", "hotValue": 12345},
      {"name": "36氪热2", "link": "https://36kr.com/p/2", "score": "678"}
    ]
  }
}
''';

      final topics = Kr36HotTopicSource.parseHotTopics(
        json,
        fetchedAt: DateTime(2026, 1, 1),
      );

      expect(topics, hasLength(2));
      expect(topics.first.source, TopicSource.kr36);
      expect(topics.first.title, '36氪热1');
      expect(topics.first.hotValue, 12345);

      expect(topics[1].title, '36氪热2');
      expect(topics[1].hotValue, 678);
    });
  });
}
