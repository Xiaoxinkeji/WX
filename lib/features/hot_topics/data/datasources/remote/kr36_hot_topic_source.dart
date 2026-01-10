import 'dart:convert';

import '../../../domain/entities/topic_source.dart';
import '../../models/hot_topic_model.dart';
import '../hot_topic_source.dart';
import 'http_fetcher.dart';
import 'parsing_utils.dart';

class Kr36HotTopicSource implements HotTopicSource {
  Kr36HotTopicSource({
    required HttpFetcher fetcher,
    Uri? hotUri,
    DateTime Function()? now,
  })  : _fetcher = fetcher,
        _hotUri = hotUri ??
            Uri.parse('https://gateway.36kr.com/api/mis/nav/home/nav/rank/hot'),
        _now = now ?? DateTime.now;

  final HttpFetcher _fetcher;
  final Uri _hotUri;
  final DateTime Function() _now;

  @override
  TopicSource get source => TopicSource.kr36;

  Map<String, String> get _headers => const {
        'Accept': 'application/json, text/plain, */*',
        'User-Agent':
            'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120 Safari/537.36',
      };

  @override
  Future<List<HotTopicModel>> fetchHotTopics() async {
    final response = await _fetcher.get(_hotUri, headers: _headers);
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw Exception('36Kr hot rank failed: ${response.statusCode}');
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

    Object? node = root?['data'];
    var data = asMap(node) ?? root;

    List<Object?>? list =
        asList(data?['hotRankList']) ?? asList(data?['items']) ?? asList(data?['list']) ?? asList(data?['data']);

    // Some endpoints wrap again.
    if (list == null) {
      final inner = asMap(data?['data']);
      list = asList(inner?['hotRankList'] ?? inner?['items'] ?? inner?['list']);
    }

    final items = list ?? const <Object?>[];

    final topics = <HotTopicModel>[];
    for (var i = 0; i < items.length; i++) {
      final item = asMap(items[i]);
      if (item == null) continue;

      final title = asString(item['title'] ?? item['name'] ?? item['word'])?.trim();
      if (title == null || title.isEmpty) continue;

      final url = tryParseUri(item['url'] ?? item['link']);
      final hotValue =
          asNum(item['hotValue'] ?? item['score'] ?? item['hot'] ?? item['hotRank']) ?? tryParseNumFromText(asString(item['hotValue']));
      final desc = asString(item['desc'] ?? item['summary'] ?? item['description']);

      topics.add(
        HotTopicModel.fromParts(
          source: TopicSource.kr36,
          rank: i + 1,
          title: title,
          url: url,
          hotValue: hotValue,
          description: desc,
          fetchedAt: fetchedAt,
        ),
      );
    }

    return topics;
  }
}
