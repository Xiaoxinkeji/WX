import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../domain/entities/topic_source.dart';
import 'hot_topics_providers.dart';
import 'hot_topics_state.dart';

class HotTopicsController extends AsyncNotifier<HotTopicsViewState> {
  @override
  FutureOr<HotTopicsViewState> build() async {
    final getHotTopics = ref.watch(getHotTopicsUseCaseProvider);
    final topics = await getHotTopics();
    return HotTopicsViewState(
      topics: topics,
      selectedSource: null,
      query: '',
      updatedAt: DateTime.now(),
    );
  }

  Future<void> setSource(TopicSource? source) async {
    final current = state.valueOrNull;
    final query = current?.query ?? '';

    state = const AsyncLoading();

    state = await AsyncValue.guard(() async {
      if (query.trim().isNotEmpty) {
        final search = ref.read(searchHotTopicsUseCaseProvider);
        final topics = await search(query, source: source);
        return HotTopicsViewState(
          topics: topics,
          selectedSource: source,
          query: query,
          updatedAt: DateTime.now(),
        );
      }

      final getHotTopics = ref.read(getHotTopicsUseCaseProvider);
      final topics = await getHotTopics(source: source);
      return HotTopicsViewState(
        topics: topics,
        selectedSource: source,
        query: '',
        updatedAt: DateTime.now(),
      );
    });
  }

  Future<void> search(String query) async {
    final current = state.valueOrNull;
    final selectedSource = current?.selectedSource;

    final normalized = query.trim();
    state = const AsyncLoading();

    state = await AsyncValue.guard(() async {
      if (normalized.isEmpty) {
        final getHotTopics = ref.read(getHotTopicsUseCaseProvider);
        final topics = await getHotTopics(source: selectedSource);
        return HotTopicsViewState(
          topics: topics,
          selectedSource: selectedSource,
          query: '',
          updatedAt: DateTime.now(),
        );
      }

      final search = ref.read(searchHotTopicsUseCaseProvider);
      final topics = await search(normalized, source: selectedSource);
      return HotTopicsViewState(
        topics: topics,
        selectedSource: selectedSource,
        query: normalized,
        updatedAt: DateTime.now(),
      );
    });
  }

  Future<void> refresh() async {
    final current = state.valueOrNull;
    final selectedSource = current?.selectedSource;
    final query = current?.query ?? '';

    state = const AsyncLoading();

    state = await AsyncValue.guard(() async {
      if (query.trim().isNotEmpty) {
        final search = ref.read(searchHotTopicsUseCaseProvider);
        final topics = await search(query, source: selectedSource, forceRefresh: true);
        return HotTopicsViewState(
          topics: topics,
          selectedSource: selectedSource,
          query: query,
          updatedAt: DateTime.now(),
        );
      }

      final refresh = ref.read(refreshHotTopicsUseCaseProvider);
      final topics = await refresh(source: selectedSource);
      return HotTopicsViewState(
        topics: topics,
        selectedSource: selectedSource,
        query: '',
        updatedAt: DateTime.now(),
      );
    });
  }
}
