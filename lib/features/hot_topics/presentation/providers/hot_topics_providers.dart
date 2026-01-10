import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../data/datasources/cache/hot_topics_cache.dart';
import '../../data/datasources/hot_topic_source.dart';
import '../../data/datasources/remote/baidu_hot_topic_source.dart';
import '../../data/datasources/remote/http_fetcher.dart';
import '../../data/datasources/remote/kr36_hot_topic_source.dart';
import '../../data/datasources/remote/weibo_hot_topic_source.dart';
import '../../data/datasources/remote/zhihu_hot_topic_source.dart';
import '../../data/repositories/hot_topics_repository_impl.dart';
import '../../domain/repositories/hot_topics_repository.dart';
import '../../domain/usecases/get_hot_topics.dart';
import '../../domain/usecases/refresh_hot_topics.dart';
import '../../domain/usecases/search_hot_topics.dart';
import 'hot_topics_controller.dart';
import 'hot_topics_state.dart';

final httpFetcherProvider = Provider<HttpFetcher>((ref) {
  return DefaultHttpFetcher();
});

final hotTopicsCacheProvider = Provider<HotTopicsCache>((ref) {
  return HotTopicsCache(ttl: const Duration(minutes: 10));
});

final hotTopicSourcesProvider = Provider<List<HotTopicSource>>((ref) {
  final fetcher = ref.watch(httpFetcherProvider);
  return [
    WeiboHotTopicSource(fetcher: fetcher),
    ZhihuHotTopicSource(fetcher: fetcher),
    BaiduHotTopicSource(fetcher: fetcher),
    Kr36HotTopicSource(fetcher: fetcher),
  ];
});

final hotTopicsRepositoryProvider = Provider<HotTopicsRepository>((ref) {
  return HotTopicsRepositoryImpl(
    sources: ref.watch(hotTopicSourcesProvider),
    cache: ref.watch(hotTopicsCacheProvider),
  );
});

final getHotTopicsUseCaseProvider = Provider<GetHotTopicsUseCase>((ref) {
  return GetHotTopicsUseCase(ref.watch(hotTopicsRepositoryProvider));
});

final refreshHotTopicsUseCaseProvider = Provider<RefreshHotTopicsUseCase>((ref) {
  return RefreshHotTopicsUseCase(ref.watch(hotTopicsRepositoryProvider));
});

final searchHotTopicsUseCaseProvider = Provider<SearchHotTopicsUseCase>((ref) {
  return SearchHotTopicsUseCase(ref.watch(hotTopicsRepositoryProvider));
});

final hotTopicsControllerProvider =
    AsyncNotifierProvider<HotTopicsController, HotTopicsViewState>(HotTopicsController.new);
