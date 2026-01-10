import 'dart:async';
import 'dart:io';

import 'package:flutter_test/flutter_test.dart';
import 'package:wechat_writing_assistant/features/hot_topics/data/datasources/remote/http_fetcher.dart';

void main() {
  test('DefaultHttpFetcher performs GET and returns status/body/headers', () async {
    final headerSeen = Completer<void>();

    final server = await HttpServer.bind(InternetAddress.loopbackIPv4, 0);
    addTearDown(() => server.close(force: true));

    server.listen((request) async {
      try {
        final header = request.headers.value('x-test');
        if (header == '1') {
          headerSeen.complete();
        } else {
          headerSeen.completeError(StateError('missing header'));
        }

        request.response.headers.set('x-response', 'ok');
        request.response.statusCode = 200;
        request.response.write('hello');
        await request.response.close();
      } catch (e, st) {
        if (!headerSeen.isCompleted) {
          headerSeen.completeError(e, st);
        }
      }
    });

    final fetcher = DefaultHttpFetcher();
    final uri = Uri.parse('http://127.0.0.1:${server.port}/');

    final response = await fetcher.get(
      uri,
      headers: const {'x-test': '1'},
      timeout: const Duration(seconds: 5),
    );

    expect(response.statusCode, 200);
    expect(response.body, contains('hello'));
    expect(response.headers['x-response'], contains('ok'));

    await headerSeen.future;
  });
}
