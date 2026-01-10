import 'dart:convert';
import 'dart:io';

import '../../../domain/entities/topic_source.dart';
import '../../models/hot_topic_model.dart';
import '../hot_topic_source.dart';
import 'http_fetcher.dart';
import 'parsing_utils.dart';

class BaiduHotTopicSource implements HotTopicSource {
  BaiduHotTopicSource({
    required HttpFetcher fetcher,
    Uri? hotUri,
    DateTime Function()? now,
  })  : _fetcher = fetcher,
        _hotUri = hotUri ??
            Uri.parse('https://top.baidu.com/api/board?platform=wise&tab=realtime'),
        _now = now ?? DateTime.now;

  final HttpFetcher _fetcher;
  final Uri _hotUri;
  final DateTime Function() _now;

  @override
  TopicSource get source => TopicSource.baidu;

  Map<String, String> get _headers => const {
        'Accept': 'application/json, text/plain, */*',
        'User-Agent':
            'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120 Safari/537.36',
      };

  @override
  Future<List<HotTopicModel>> fetchHotTopics() async {
    final response = await _fetcher.get(_hotUri, headers: _headers);
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw HttpException('Baidu hot board failed: ${response.statusCode}');
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
    final data = asMap(root?['data']);

    final cards = asList(data?['cards']) ?? const <Object?>[];

    final entries = <Map<String, Object?>>[];
    for (final card in cards) {
      final cardMap = asMap(card);
      if (cardMap == null) continue;

      final content = asList(cardMap['content']);
      if (content == null) continue;

      for (final item in content) {
        final m = asMap(item);
        if (m != null) entries.add(m);
      }
    }

    final topics = <HotTopicModel>[];
    for (var i = 0; i < entries.length; i++) {
      final item = entries[i];
      final title = asString(item['word'] ?? item['keyword'] ?? item['title'])?.trim();
      if (title == null || title.isEmpty) continue;

      final url = tryParseUri(item['url'] ?? item['link']);
      final hotValue = asNum(item['hotScore'] ?? item['hot_score'] ?? item['hotValue'] ?? item['score']) ??
          tryParseNumFromText(asString(item['hotScore']));
      final desc = asString(item['desc'] ?? item['desc1'] ?? item['summary']);

      topics.add(
        HotTopicModel.fromParts(
          source: TopicSource.baidu,
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
