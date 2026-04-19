import 'package:mobile/core/constants/api_constants.dart';
import 'package:mobile/core/network/api_client.dart';
import 'package:mobile/domain/models/user_model.dart';

typedef AuthResult = ({
  String accessToken,
  String refreshToken,
  UserModel user,
});

typedef RefreshResult = ({
  String accessToken,
  String refreshToken,
});

class AuthRemote {
  final ApiClient _client;

  AuthRemote(this._client);

  Future<AuthResult> register({
    required String businessName,
    required String phone,
    required String password,
  }) async {
    final response = await _client.dio.post(
      ApiConstants.register,
      data: {
        'business_name': businessName,
        'phone': phone,
        'password': password,
      },
    );

    final data = response.data['data'] as Map<String, dynamic>;
    return (
      accessToken: data['access_token'] as String,
      refreshToken: data['refresh_token'] as String,
      user: UserModel.fromJson(data['user'] as Map<String, dynamic>),
    );
  }

  Future<AuthResult> login({
    required String phone,
    required String password,
  }) async {
    final response = await _client.dio.post(
      ApiConstants.login,
      data: {'phone': phone, 'password': password},
    );

    final data = response.data['data'] as Map<String, dynamic>;
    return (
      accessToken: data['access_token'] as String,
      refreshToken: data['refresh_token'] as String,
      user: UserModel.fromJson(data['user'] as Map<String, dynamic>),
    );
  }

  Future<RefreshResult> refresh({required String refreshToken}) async {
    final response = await _client.dio.post(
      ApiConstants.refresh,
      data: {'refresh_token': refreshToken},
    );

    final data = response.data['data'] as Map<String, dynamic>;
    return (
      accessToken: data['access_token'] as String,
      refreshToken: data['refresh_token'] as String,
    );
  }

  Future<void> logout({required String refreshToken}) async {
    await _client.dio.post(
      ApiConstants.logout,
      data: {'refresh_token': refreshToken},
    );
  }
}
