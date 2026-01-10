import '../../domain/entities/hot_topic.dart';
import '../../domain/entities/topic_source.dart';

class HotTopicsViewState {
  const HotTopicsViewState({
    required this.topics,
    required this.selectedSource,
    required this.query,
    required this.updatedAt,
  });

  final List<HotTopic> topics;
  final TopicSource? selectedSource;
  final String query;
  final DateTime updatedAt;

  HotTopicsViewState copyWith({
    List<HotTopic>? topics,
    TopicSource? selectedSource,
    String? query,
    DateTime? updatedAt,
  }) {
    return HotTopicsViewState(
      topics: topics ?? this.topics,
      selectedSource: selectedSource ?? this.selectedSource,
      query: query ?? this.query,
      updatedAt: updatedAt ?? this.updatedAt,
    );
  }
}
