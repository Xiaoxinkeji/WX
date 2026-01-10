import 'package:flutter_test/flutter_test.dart';
import 'package:wechat_writing_assistant/features/hot_topics/data/datasources/remote/parsing_utils.dart';

void main() {
  group('parsing_utils', () {
    test('asString/asInt/asNum handle different inputs', () {
      expect(asString(1), '1');
      expect(asString(null), isNull);

      expect(asInt(1), 1);
      expect(asInt(1.9), 1);
      expect(asInt('42'), 42);
      expect(asInt('x'), isNull);

      expect(asNum(1), 1);
      expect(asNum('3.14'), 3.14);
      expect(asNum('x'), isNull);
    });

    test('asMap/asList convert dynamic collections', () {
      final map = asMap({'a': 1, 2: 'ignored'});
      expect(map, isNotNull);
      expect(map!['a'], 1);
      expect(map.containsKey('2'), isFalse);

      final list = asList([1, 2, 3]);
      expect(list, isNotNull);
      expect(list, hasLength(3));
    });

    test('tryParseNumFromText extracts first number', () {
      expect(tryParseNumFromText('123 万热度'), 123);
      expect(tryParseNumFromText('1,234'), 1234);
      expect(tryParseNumFromText('nope'), isNull);
    });

    test('tryParseUri handles different types', () {
      expect(tryParseUri('https://example.com')!.host, 'example.com');
      expect(tryParseUri(Uri.parse('https://example.com'))!.host, 'example.com');
      expect(tryParseUri(null), isNull);
    });

    test('containsIgnoreCase normalizes query', () {
      expect(containsIgnoreCase('Flutter Riverpod', ' riverPOD '), isTrue);
      expect(containsIgnoreCase('Flutter', 'Dart'), isFalse);
    });
  });
}
