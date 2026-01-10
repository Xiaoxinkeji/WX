import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:wechat_writing_assistant/features/hot_topics/data/datasources/remote/http_fetcher.dart';

class MockHttpClient extends http.BaseClient {
  final Map<String, String> requestHeaders = {};
  final http.Response response;

  MockHttpClient(this.response);

  @override
  Future<http.StreamedResponse> send(http.BaseRequest request) async {
    requestHeaders.addAll(request.headers);
    return http.StreamedResponse(
      Stream.value(response.bodyBytes),
      response.statusCode,
      headers: response.headers,
    );
  }
}

void main() {
  test('DefaultHttpFetcher performs GET and returns status/body/headers', () async {
    final mockResponse = http.Response(
      'hello',
      200,
      headers: {'x-response': 'ok'},
    );
    final mockClient = MockHttpClient(mockResponse);
    final fetcher = DefaultHttpFetcher(client: mockClient);

    final response = await fetcher.get(
      Uri.parse('http://example.com/'),
      headers: const {'x-test': '1'},
      timeout: const Duration(seconds: 5),
    );

    expect(response.statusCode, 200);
    expect(response.body, contains('hello'));
    expect(response.headers['x-response'], 'ok');
    expect(mockClient.requestHeaders['x-test'], '1');
  });
}
