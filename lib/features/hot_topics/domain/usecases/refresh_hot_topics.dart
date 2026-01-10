import '../entities/hot_topic.dart';
import '../entities/topic_source.dart';
import '../repositories/hot_topics_repository.dart';

class RefreshHotTopicsUseCase {
  const RefreshHotTopicsUseCase(this._repository);

  final HotTopicsRepository _repository;

  Future<List<HotTopic>> call({TopicSource? source}) {
    return _repository.refreshHotTopics(source: source);
  }
}
