/// Hot topic source platform.
///
/// Note: Domain layer should stay UI-framework agnostic.
enum TopicSource {
  weibo,
  zhihu,
  baidu,
  kr36;

  String get label => switch (this) {
        TopicSource.weibo => '微博',
        TopicSource.zhihu => '知乎',
        TopicSource.baidu => '百度',
        TopicSource.kr36 => '36氪',
      };

  /// Stable identifier for persistence/cache keys.
  String get key => name;

  Uri? get homepage => switch (this) {
        TopicSource.weibo => Uri.tryParse('https://weibo.com/'),
        TopicSource.zhihu => Uri.tryParse('https://www.zhihu.com/'),
        TopicSource.baidu => Uri.tryParse('https://top.baidu.com/'),
        TopicSource.kr36 => Uri.tryParse('https://36kr.com/'),
      };
}
