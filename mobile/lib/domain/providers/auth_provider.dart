import 'package:riverpod_annotation/riverpod_annotation.dart';
import 'package:mobile/core/network/api_client.dart';
import 'package:mobile/data/local/secure_storage.dart';
import 'package:mobile/data/remote/auth_remote.dart';
import 'package:mobile/domain/models/user_model.dart';

part 'auth_provider.g.dart';

@riverpod
ApiClient apiClient(ApiClientRef ref) => ApiClient();

@riverpod
AuthRemote authRemote(AuthRemoteRef ref) =>
    AuthRemote(ref.watch(apiClientProvider));

@riverpod
class AuthNotifier extends _$AuthNotifier {
  @override
  Future<UserModel?> build() async {
    final userData = await SecureStorage.getUserData();
    if (userData == null) return null;
    return UserModel.fromJson(userData);
  }

  Future<void> register({
    required String businessName,
    required String phone,
    required String password,
  }) async {
    state = const AsyncLoading();
    try {
      final result = await ref.read(authRemoteProvider).register(
            businessName: businessName,
            phone: phone,
            password: password,
          );
      await _persistSession(result.accessToken, result.refreshToken, result.user);
      state = AsyncData(result.user);
    } catch (e, st) {
      state = AsyncError(e, st);
      rethrow;
    }
  }

  Future<void> login({
    required String phone,
    required String password,
  }) async {
    state = const AsyncLoading();
    try {
      final result = await ref.read(authRemoteProvider).login(
            phone: phone,
            password: password,
          );
      await _persistSession(result.accessToken, result.refreshToken, result.user);
      state = AsyncData(result.user);
    } catch (e, st) {
      state = AsyncError(e, st);
      rethrow;
    }
  }

  Future<void> logout() async {
    final refreshToken = await SecureStorage.getRefreshToken();
    if (refreshToken != null) {
      try {
        await ref.read(authRemoteProvider).logout(refreshToken: refreshToken);
      } catch (_) {
        // Lanjut logout lokal meski request gagal
      }
    }
    await SecureStorage.clearAll();
    state = const AsyncData(null);
  }

  bool get isLoggedIn => state.valueOrNull != null;

  UserModel? get currentUser => state.valueOrNull;

  Future<void> _persistSession(
    String accessToken,
    String refreshToken,
    UserModel user,
  ) async {
    await SecureStorage.saveAccessToken(accessToken);
    await SecureStorage.saveRefreshToken(refreshToken);
    await SecureStorage.saveUserData(user.toJson());
  }
}
