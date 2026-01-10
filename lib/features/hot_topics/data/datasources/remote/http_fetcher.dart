import 'dart:async';
import 'package:http/http.dart' as http;

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
  DefaultHttpFetcher({http.Client? client}) : _client = client ?? http.Client();

  final http.Client _client;

  @override
  Future<FetchResponse> get(
    Uri uri, {
    Map<String, String> headers = const {},
    Duration timeout = const Duration(seconds: 15),
  }) async {
    final response = await _client
        .get(uri, headers: headers)
        .timeout(timeout);

    return FetchResponse(
      statusCode: response.statusCode,
      body: response.body,
      headers: response.headers,
    );
  }
}
