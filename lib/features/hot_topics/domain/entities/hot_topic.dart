import 'topic_source.dart';

class HotTopic {
  const HotTopic({
    required this.id,
    required this.source,
    required this.rank,
    required this.title,
    this.url,
    this.hotValue,
    this.description,
    required this.fetchedAt,
  });

  final String id;
  final TopicSource source;

  /// 1-based rank within a platform list.
  final int rank;

  final String title;
  final Uri? url;

  /// Platform-specific popularity metric.
  final num? hotValue;

  /// Optional short description / excerpt.
  final String? description;

  /// When this topic list item was fetched.
  final DateTime fetchedAt;

  HotTopic copyWith({
    String? id,
    TopicSource? source,
    int? rank,
    String? title,
    Uri? url,
    num? hotValue,
    String? description,
    DateTime? fetchedAt,
  }) {
    return HotTopic(
      id: id ?? this.id,
      source: source ?? this.source,
      rank: rank ?? this.rank,
      title: title ?? this.title,
      url: url ?? this.url,
      hotValue: hotValue ?? this.hotValue,
      description: description ?? this.description,
      fetchedAt: fetchedAt ?? this.fetchedAt,
    );
  }

  @override
  String toString() {
    return 'HotTopic(id: $id, source: ${source.key}, rank: $rank, title: $title)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other is HotTopic &&
            other.id == id &&
            other.source == source &&
            other.rank == rank &&
            other.title == title &&
            other.url == url &&
            other.hotValue == hotValue &&
            other.description == description &&
            other.fetchedAt == fetchedAt);
  }

  @override
  int get hashCode => Object.hash(
        id,
        source,
        rank,
        title,
        url,
        hotValue,
        description,
        fetchedAt,
      );
}
