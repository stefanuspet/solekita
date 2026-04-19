import 'package:dio/dio.dart';
import 'package:mobile/core/constants/api_constants.dart';
import 'package:mobile/data/local/secure_storage.dart';

import 'api_exception.dart';

class ApiClient {
  final Dio _dio;
  final Future<String?> Function() _getAccessToken;
  final Future<String?> Function() _getRefreshToken;
  final Future<void> Function(String) _saveAccessToken;

  ApiClient({
    String? baseUrl,
    Future<String?> Function()? getAccessToken,
    Future<String?> Function()? getRefreshToken,
    Future<void> Function(String)? saveAccessToken,
  })  : _getAccessToken = getAccessToken ?? SecureStorage.getAccessToken,
        _getRefreshToken = getRefreshToken ?? SecureStorage.getRefreshToken,
        _saveAccessToken = saveAccessToken ?? SecureStorage.saveAccessToken,
        _dio = Dio(
          BaseOptions(
            baseUrl: baseUrl ?? ApiConstants.baseUrl ?? '',
            connectTimeout: const Duration(seconds: 10),
            receiveTimeout: const Duration(seconds: 30),
            headers: {'Content-Type': 'application/json'},
          ),
        ) {
    _dio.interceptors.add(_authInterceptor());
    _dio.interceptors.add(_refreshInterceptor());
    _dio.interceptors.add(_errorInterceptor());
  }

  Dio get dio => _dio;

  Interceptor _authInterceptor() {
    return InterceptorsWrapper(
      onRequest: (options, handler) async {
        final token = await _getAccessToken();
        if (token != null) {
          options.headers['Authorization'] = 'Bearer $token';
        }
        handler.next(options);
      },
    );
  }

  Interceptor _refreshInterceptor() {
    return InterceptorsWrapper(
      onError: (error, handler) async {
        if (error.response?.statusCode != 401) {
          return handler.next(error);
        }

        try {
          final refreshToken = await _getRefreshToken();
          if (refreshToken == null) {
            return handler.next(error);
          }

          final response = await _dio.post(
            ApiConstants.refresh,
            data: {'refresh_token': refreshToken},
          );

          final newAccessToken = response.data['data']['access_token'] as String;
          await _saveAccessToken(newAccessToken);

          final retryOptions = error.requestOptions;
          retryOptions.headers['Authorization'] = 'Bearer $newAccessToken';
          final retryResponse = await _dio.fetch(retryOptions);
          return handler.resolve(retryResponse);
        } catch (_) {
          return handler.next(error);
        }
      },
    );
  }

  Interceptor _errorInterceptor() {
    return InterceptorsWrapper(
      onError: (error, handler) {
        final response = error.response;
        if (response != null) {
          final data = response.data is Map<String, dynamic>
              ? response.data as Map<String, dynamic>
              : <String, dynamic>{'message': 'Terjadi kesalahan'};
          final apiException = ApiException.fromJson(
            response.statusCode ?? 500,
            data,
          );
          return handler.reject(
            DioException(
              requestOptions: error.requestOptions,
              error: apiException,
            ),
          );
        }
        return handler.next(error);
      },
    );
  }
}
