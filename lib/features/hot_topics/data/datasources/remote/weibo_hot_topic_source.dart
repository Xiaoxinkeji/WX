import 'dart:convert';
import 'dart:io';

import '../../../domain/entities/topic_source.dart';
import '../../models/hot_topic_model.dart';
import '../hot_topic_source.dart';
import 'http_fetcher.dart';
import 'parsing_utils.dart';

class WeiboHotTopicSource implements HotTopicSource {
  WeiboHotTopicSource({
    required HttpFetcher fetcher,
    Uri? hotUri,
    DateTime Function()? now,
  })  : _fetcher = fetcher,
        _hotUri = hotUri ?? Uri.parse('https://weibo.com/ajax/side/hotSearch'),
        _now = now ?? DateTime.now;

  final HttpFetcher _fetcher;
  final Uri _hotUri;
  final DateTime Function() _now;

  @override
  TopicSource get source => TopicSource.weibo;

  Map<String, String> get _headers => const {
        'Accept': 'application/json, text/plain, */*',
        'User-Agent':
            'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120 Safari/537.36',
      };

  @override
  Future<List<HotTopicModel>> fetchHotTopics() async {
    final response = await _fetcher.get(_hotUri, headers: _headers);
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw HttpException('Weibo hotSearch failed: ${response.statusCode}');
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
    final realtime = asList(data?['realtime']) ?? const <Object?>[];

    final topics = <HotTopicModel>[];
    for (var i = 0; i < realtime.length; i++) {
      final item = asMap(realtime[i]);
      if (item == null) continue;

      final title = asString(item['note'] ?? item['word'])?.trim();
      if (title == null || title.isEmpty) continue;

      final rank = asInt(item['rank'] ?? item['realpos'] ?? item['num']) ?? (i + 1);
      final hotValue = asNum(item['raw_hot'] ?? item['rawHot'] ?? item['num'] ?? item['hot']) ??
          tryParseNumFromText(asString(item['raw_hot']));

      final url = tryParseUri(item['link'] ?? item['url']);

      topics.add(
        HotTopicModel.fromParts(
          source: TopicSource.weibo,
          rank: rank,
          title: title,
          url: url,
          hotValue: hotValue,
          fetchedAt: fetchedAt,
        ),
      );
    }

    topics.sort((a, b) => a.rank.compareTo(b.rank));
    return topics;
  }
}
