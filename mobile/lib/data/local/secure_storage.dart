import 'dart:convert';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

class SecureStorage {
  static const FlutterSecureStorage _storage = FlutterSecureStorage();

  static const String _accessTokenKey = 'access_token';
  static const String _refreshTokenKey = 'refresh_token';
  static const String _userDataKey = 'user_data';

  // Access token
  static Future<void> saveAccessToken(String token) async {
    await _storage.write(key: _accessTokenKey, value: token);
  }

  static Future<String?> getAccessToken() async {
    return await _storage.read(key: _accessTokenKey);
  }

  static Future<void> deleteAccessToken() async {
    await _storage.delete(key: _accessTokenKey);
  }

  // Refresh token
  static Future<void> saveRefreshToken(String token) async {
    await _storage.write(key: _refreshTokenKey, value: token);
  }

  static Future<String?> getRefreshToken() async {
    return await _storage.read(key: _refreshTokenKey);
  }

  static Future<void> deleteRefreshToken() async {
    await _storage.delete(key: _refreshTokenKey);
  }

  // User data (JSON string)
  static Future<void> saveUserData(Map<String, dynamic> user) async {
    await _storage.write(key: _userDataKey, value: jsonEncode(user));
  }

  static Future<Map<String, dynamic>?> getUserData() async {
    final jsonString = await _storage.read(key: _userDataKey);
    if (jsonString == null) return null;
    return jsonDecode(jsonString) as Map<String, dynamic>;
  }

  static Future<void> deleteUserData() async {
    await _storage.delete(key: _userDataKey);
  }

  // First-time setup flag
  static const String _setupCompletedKey = 'first_setup_completed';

  static Future<void> setSetupCompleted() async {
    await _storage.write(key: _setupCompletedKey, value: 'true');
  }

  static Future<bool> isSetupCompleted() async {
    final value = await _storage.read(key: _setupCompletedKey);
    return value == 'true';
  }

  // Hapus semua (logout)
  static Future<void> clearAll() async {
    await _storage.deleteAll();
  }
}
