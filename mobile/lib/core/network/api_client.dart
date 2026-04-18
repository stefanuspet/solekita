import 'package:dio/dio.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:mobile/core/constants/api_constants.dart';

import 'api_exception.dart';

class ApiClient {
  final Dio _dio;
  final FlutterSecureStorage _storage;

  ApiClient()
      : _dio = Dio(
          BaseOptions(
            baseUrl: ApiConstants.baseUrl ?? '',
            connectTimeout: const Duration(seconds: 10),
            receiveTimeout: const Duration(seconds: 30),
            headers: {
              'Content-Type': 'application/json',
            },
          ),
        ),
        _storage = const FlutterSecureStorage() {
    _dio.interceptors.add(_authInterceptor());
    _dio.interceptors.add(_refreshInterceptor());
    _dio.interceptors.add(_errorInterceptor());
  }

  Dio get dio => _dio;

  // ========================
  // AUTH INTERCEPTOR
  // ========================
  Interceptor _authInterceptor() {
    return InterceptorsWrapper(
      onRequest: (options, handler) async {
        final token = await _storage.read(key: 'access_token');

        if (token != null) {
          options.headers['Authorization'] = 'Bearer $token';
        }

        handler.next(options);
      },
    );
  }

  // ========================
  // REFRESH TOKEN INTERCEPTOR
  // ========================
  Interceptor _refreshInterceptor() {
    return InterceptorsWrapper(
      onError: (error, handler) async {
        // kalau bukan 401 → lanjut
        if (error.response?.statusCode != 401) {
          return handler.next(error);
        }

        try {
          final refreshToken = await _storage.read(key: 'refresh_token');

          if (refreshToken == null) {
            return handler.next(error);
          }

          // request refresh token
          final response = await _dio.post(
            ApiConstants.refresh,
            data: {'refresh_token': refreshToken},
          );

          final newAccessToken = response.data['access_token'];

          // simpan token baru
          await _storage.write(key: 'access_token', value: newAccessToken);

          // retry request sebelumnya
          final requestOptions = error.requestOptions;
          requestOptions.headers['Authorization'] = 'Bearer $newAccessToken';

          final cloneReq = await _dio.fetch(requestOptions);

          return handler.resolve(cloneReq);
        } catch (e) {
          return handler.next(error);
        }
      },
    );
  }

  // ========================
  // ERROR INTERCEPTOR
  // ========================
  Interceptor _errorInterceptor() {
    return InterceptorsWrapper(
      onError: (error, handler) {
        if (error is DioException) {
          final response = error.response;

          if (response != null) {
            final apiException = ApiException.fromJson(
              response.statusCode ?? 500,
              response.data,
            );

            return handler.reject(
              DioException(
                requestOptions: error.requestOptions,
                error: apiException,
              ),
            );
          }
        }

        return handler.next(error);
      },
    );
  }
}
