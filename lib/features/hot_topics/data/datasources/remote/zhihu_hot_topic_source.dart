import 'dart:convert';

import '../../../domain/entities/topic_source.dart';
import '../../models/hot_topic_model.dart';
import '../hot_topic_source.dart';
import 'http_fetcher.dart';
import 'parsing_utils.dart';

class ZhihuHotTopicSource implements HotTopicSource {
  ZhihuHotTopicSource({
    required HttpFetcher fetcher,
    Uri? hotUri,
    DateTime Function()? now,
  })  : _fetcher = fetcher,
        _hotUri = hotUri ??
            Uri.parse(
              'https://www.zhihu.com/api/v3/feed/topstory/hot-lists/total?limit=50&desktop=true',
            ),
        _now = now ?? DateTime.now;

  final HttpFetcher _fetcher;
  final Uri _hotUri;
  final DateTime Function() _now;

  @override
  TopicSource get source => TopicSource.zhihu;

  Map<String, String> get _headers => const {
        'Accept': 'application/json, text/plain, */*',
        'User-Agent':
            'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120 Safari/537.36',
      };

  @override
  Future<List<HotTopicModel>> fetchHotTopics() async {
    final response = await _fetcher.get(_hotUri, headers: _headers);
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw Exception('Zhihu hot list failed: ${response.statusCode}');
    }
    return parseHotTopics(response.body, fetchedAt: _now());
  }

  @override
  Future<List<HotTopicModel>> search(String query) async {
    final trimmed = query.trim();
    if (trimmed.isEmpty) return fetchHotTopics();

    final topics = await fetchHotTopics();
    return topics.where((t) => containsIgnoreCase(t.title, trimmed)).toList(growable: false);
  }

  static List<HotTopicModel> parseHotTopics(
    String body, {
    required DateTime fetchedAt,
  }) {
    final decoded = jsonDecode(body);
    final root = asMap(decoded);
    final items = asList(root?['data']) ?? const <Object?>[];

    final topics = <HotTopicModel>[];
    for (var i = 0; i < items.length; i++) {
      final item = asMap(items[i]);
      if (item == null) continue;

      final target = asMap(item['target']);

      final title = asString(target?['title'] ?? item['title'])?.trim();
      if (title == null || title.isEmpty) continue;

      final excerpt = asString(target?['excerpt'] ?? target?['excerpt_new'] ?? target?['description']);

      final url = tryParseUri(target?['url'] ?? target?['url_token'] ?? target?['urlToken'] ?? item['url']);

      final hotValue =
          tryParseNumFromText(asString(item['detail_text'] ?? item['detailText'] ?? item['heat']));

      topics.add(
        HotTopicModel.fromParts(
          source: TopicSource.zhihu,
          rank: i + 1,
          title: title,
          url: url,
          hotValue: hotValue,
          description: excerpt,
          fetchedAt: fetchedAt,
        ),
      );
    }

    return topics;
  }
}
