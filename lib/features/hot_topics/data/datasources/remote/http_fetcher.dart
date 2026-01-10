import 'dart:async';
import 'dart:convert';
import 'dart:io';

class FetchResponse {
  const FetchResponse({
    required this.statusCode,
    required this.body,
    required this.headers,
  });

  final int statusCode;
  final String body;
  final Map<String, String> headers;
}

abstract class HttpFetcher {
  Future<FetchResponse> get(
    Uri uri, {
    Map<String, String> headers = const {},
    Duration timeout = const Duration(seconds: 15),
  });
}

class DefaultHttpFetcher implements HttpFetcher {
  DefaultHttpFetcher({HttpClient? client}) : _client = client ?? HttpClient();

  final HttpClient _client;

  @override
  Future<FetchResponse> get(
    Uri uri, {
    Map<String, String> headers = const {},
    Duration timeout = const Duration(seconds: 15),
  }) async {
    final request = await _client.getUrl(uri).timeout(timeout);
    headers.forEach(request.headers.set);

    final response = await request.close().timeout(timeout);
    final status = response.statusCode;

    final bytes = await response.fold<List<int>>(<int>[], (buffer, chunk) {
      buffer.addAll(chunk);
      return buffer;
    });
    final body = utf8.decode(bytes, allowMalformed: true);

    final responseHeaders = <String, String>{};
    response.headers.forEach((name, values) {
      if (values.isNotEmpty) responseHeaders[name] = values.join(',');
    });

    return FetchResponse(statusCode: status, body: body, headers: responseHeaders);
  }
}
