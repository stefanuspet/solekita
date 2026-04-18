import 'dart:convert';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

class SecureStorage {
  static const FlutterSecureStorage _storage = FlutterSecureStorage();

  // ========================
  // KEYS
  // ========================
  static const String _accessTokenKey = 'access_token';
  static const String _refreshTokenKey = 'refresh_token';
  static const String _userDataKey = 'user_data';

  // ========================
  // ACCESS TOKEN
  // ========================
  static Future<void> saveAccessToken(String token) async {
    await _storage.write(key: _accessTokenKey, value: token);
  }

  static Future<String?> getAccessToken() async {
    return await _storage.read(key: _accessTokenKey);
  }

  static Future<void> deleteAccessToken() async {
    await _storage.delete(key: _accessTokenKey);
  }

  // ========================
  // REFRESH TOKEN
  // ========================
  static Future<void> saveRefreshToken(String token) async {
    await _storage.write(key: _refreshTokenKey, value: token);
  }

  static Future<String?> getRefreshToken() async {
    return await _storage.read(key: _refreshTokenKey);
  }

  static Future<void> deleteRefreshToken() async {
    await _storage.delete(key: _refreshTokenKey);
  }

  // ========================
  // USER DATA (JSON STRING)
  // ========================
  static Future<void> saveUserData(Map<String, dynamic> user) async {
    final jsonString = jsonEncode(user);
    await _storage.write(key: _userDataKey, value: jsonString);
  }

  static Future<Map<String, dynamic>?> getUserData() async {
    final jsonString = await _storage.read(key: _userDataKey);
    if (jsonString == null) return null;

    return jsonDecode(jsonString);
  }

  static Future<void> deleteUserData() async {
    await _storage.delete(key: _userDataKey);
  }

  // ========================
  // CLEAR ALL (LOGOUT)
  // ========================
  static Future<void> clearAll() async {
    await _storage.deleteAll();
  }
}
