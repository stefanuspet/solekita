import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mobile/core/network/api_client.dart';
import 'package:mobile/core/network/api_exception.dart';

// ── Mock HTTP adapter ─────────────────────────────────────────────────────────

typedef _Handler = Future<ResponseBody> Function(RequestOptions);

class _MockAdapter implements HttpClientAdapter {
  _Handler handler;
  _MockAdapter(this.handler);

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) =>
      handler(options);

  @override
  void close({bool force = false}) {}
}

ResponseBody _ok(Map<String, dynamic> body) {
  return ResponseBody.fromBytes(
    utf8.encode(jsonEncode(body)),
    200,
    headers: {Headers.contentTypeHeader: [Headers.jsonContentType]},
  );
}

ResponseBody _error(int status, Map<String, dynamic> body) {
  return ResponseBody.fromBytes(
    utf8.encode(jsonEncode(body)),
    status,
    headers: {Headers.contentTypeHeader: [Headers.jsonContentType]},
  );
}

// ── Helper: buat ApiClient dengan in-memory storage & mock HTTP ───────────────

ApiClient _makeClient({
  Map<String, String?>? store,
  required _Handler httpHandler,
}) {
  final storage = store ?? {};
  final client = ApiClient(
    baseUrl: 'http://localhost',
    getAccessToken: () async => storage['access_token'],
    getRefreshToken: () async => storage['refresh_token'],
    saveAccessToken: (t) async => storage['access_token'] = t,
  );
  client.dio.httpClientAdapter = _MockAdapter(httpHandler);
  return client;
}

// ── Tests ─────────────────────────────────────────────────────────────────────

void main() {
  group('AuthInterceptor', () {
    test('inject Authorization header jika ada access token', () async {
      RequestOptions? captured;

      final client = _makeClient(
        store: {'access_token': 'tok-abc'},
        httpHandler: (opts) async {
          captured = opts;
          return _ok({'success': true});
        },
      );

      await client.dio.get('/health');

      expect(captured?.headers['Authorization'], equals('Bearer tok-abc'));
    });

    test('tidak inject header jika tidak ada access token', () async {
      RequestOptions? captured;

      final client = _makeClient(
        store: {},
        httpHandler: (opts) async {
          captured = opts;
          return _ok({'success': true});
        },
      );

      await client.dio.get('/health');

      expect(captured?.headers.containsKey('Authorization'), isFalse);
    });
  });

  group('RefreshInterceptor', () {
    test('auto refresh token dan retry request saat 401', () async {
      final storage = <String, String?>{
        'access_token': 'old-token',
        'refresh_token': 'ref-tok',
      };

      int callCount = 0;

      final client = _makeClient(
        store: storage,
        httpHandler: (opts) async {
          callCount++;

          // 1: request asli → 401
          if (callCount == 1) {
            return _error(401, {'success': false, 'message': 'Unauthorized'});
          }

          // 2: POST /auth/refresh → 200 dengan token baru
          if (opts.path.contains('/auth/refresh')) {
            return _ok({
              'success': true,
              'data': {'access_token': 'new-token', 'refresh_token': 'ref-tok'},
            });
          }

          // 3: retry request asli → 200
          return _ok({'success': true, 'data': 'ok'});
        },
      );

      final response = await client.dio.get('/orders');

      expect(response.statusCode, equals(200));
      expect(storage['access_token'], equals('new-token'));
      expect(callCount, equals(3)); // asli + refresh + retry
    });

    test('lanjutkan error jika tidak ada refresh token', () async {
      final client = _makeClient(
        store: {'access_token': 'tok'},
        httpHandler: (opts) async =>
            _error(401, {'success': false, 'message': 'Unauthorized'}),
      );

      expect(
        () => client.dio.get('/orders'),
        throwsA(isA<DioException>()),
      );
    });
  });

  group('ErrorInterceptor', () {
    test('konversi DioException ke ApiException saat response error', () async {
      final client = _makeClient(
        httpHandler: (opts) async => _error(422, {
          'success': false,
          'message': 'Proses gagal',
        }),
      );

      DioException? caught;
      try {
        await client.dio.get('/orders');
      } on DioException catch (e) {
        caught = e;
      }

      expect(caught, isNotNull);
      expect(caught!.error, isA<ApiException>());

      final apiEx = caught.error as ApiException;
      expect(apiEx.statusCode, equals(422));
      expect(apiEx.message, equals('Proses gagal'));
    });

    test('tetap DioException jika tidak ada response body (network error)',
        () async {
      final client = ApiClient(
        baseUrl: 'http://0.0.0.0:1',
        getAccessToken: () async => null,
        getRefreshToken: () async => null,
        saveAccessToken: (_) async {},
      );
      client.dio.options.connectTimeout = const Duration(milliseconds: 200);

      expect(
        () => client.dio.get('/health'),
        throwsA(isA<DioException>()),
      );
    });
  });
}
