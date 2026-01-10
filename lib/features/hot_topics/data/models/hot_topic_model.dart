import '../../domain/entities/hot_topic.dart';
import '../../domain/entities/topic_source.dart';

class HotTopicModel extends HotTopic {
  HotTopicModel({
    required super.id,
    required super.source,
    required super.rank,
    required super.title,
    required super.fetchedAt,
    super.url,
    super.hotValue,
    super.description,
  });

  factory HotTopicModel.fromParts({
    required TopicSource source,
    required int rank,
    required String title,
    Uri? url,
    num? hotValue,
    String? description,
    required DateTime fetchedAt,
  }) {
    return HotTopicModel(
      id: generateId(source: source, title: title, url: url),
      source: source,
      rank: rank,
      title: title,
      url: url,
      hotValue: hotValue,
      description: description,
      fetchedAt: fetchedAt,
    );
  }

  static String generateId({
    required TopicSource source,
    required String title,
    Uri? url,
  }) {
    final normalizedTitle = title.trim().toLowerCase().replaceAll(RegExp(r'\s+'), ' ');
    final key = (url?.toString().isNotEmpty ?? false) ? url.toString() : normalizedTitle;
    return '${source.key}:$key';
  }

  Map<String, Object?> toMap() {
    return {
      'id': id,
      'source': source.key,
      'rank': rank,
      'title': title,
      'url': url?.toString(),
      'hotValue': hotValue,
      'description': description,
      'fetchedAt': fetchedAt.toIso8601String(),
    };
  }

  factory HotTopicModel.fromMap(Map<String, Object?> map) {
    final sourceKey = map['source'] as String?;
    final source = TopicSource.values.firstWhere(
      (s) => s.key == sourceKey,
      orElse: () => TopicSource.weibo,
    );

    return HotTopicModel(
      id: map['id'] as String,
      source: source,
      rank: map['rank'] as int,
      title: map['title'] as String,
      url: (map['url'] as String?) == null ? null : Uri.tryParse(map['url'] as String),
      hotValue: map['hotValue'] as num?,
      description: map['description'] as String?,
      fetchedAt: DateTime.parse(map['fetchedAt'] as String),
    );
  }
}
