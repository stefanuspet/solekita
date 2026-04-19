import 'package:mobile/core/constants/api_constants.dart';
import 'package:mobile/core/network/api_client.dart';

class StaffModel {
  final String id;
  final String name;
  final String phone;
  final bool isOwner;
  final bool isActive;
  final List<String> permissions;
  final DateTime? lastLoginAt;
  final DateTime? createdAt;

  const StaffModel({
    required this.id,
    required this.name,
    required this.phone,
    required this.isOwner,
    required this.isActive,
    required this.permissions,
    this.lastLoginAt,
    this.createdAt,
  });

  factory StaffModel.fromJson(Map<String, dynamic> json) => StaffModel(
        id: json['id'] as String,
        name: json['name'] as String,
        phone: json['phone'] as String,
        isOwner: json['is_owner'] as bool? ?? false,
        isActive: json['is_active'] as bool,
        permissions: (json['permissions'] as List<dynamic>)
            .map((e) => e as String)
            .toList(),
        lastLoginAt: json['last_login_at'] != null
            ? DateTime.parse(json['last_login_at'] as String)
            : null,
        createdAt: json['created_at'] != null
            ? DateTime.parse(json['created_at'] as String)
            : null,
      );
}

typedef CreateUserResult = ({StaffModel user, String temporaryPassword});

class UserRemote {
  final ApiClient _client;

  UserRemote(this._client);

  Future<List<StaffModel>> listUsers({bool? isActive}) async {
    final response = await _client.dio.get(
      ApiConstants.users,
      queryParameters: {
        if (isActive != null) 'is_active': isActive,
      },
    );
    return (response.data['data'] as List<dynamic>)
        .map((e) => StaffModel.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<CreateUserResult> createUser({
    required String name,
    required String phone,
    required List<String> permissions,
  }) async {
    final response = await _client.dio.post(
      ApiConstants.users,
      data: {'name': name, 'phone': phone, 'permissions': permissions},
    );
    final data = response.data['data'] as Map<String, dynamic>;
    return (
      user: StaffModel.fromJson(data),
      temporaryPassword: data['temporary_password'] as String,
    );
  }

  Future<StaffModel> updateUser(
    String id, {
    String? name,
    List<String>? permissions,
    bool? isActive,
  }) async {
    final response = await _client.dio.patch(
      '${ApiConstants.users}/$id',
      data: {
        if (name != null) 'name': name,
        if (permissions != null) 'permissions': permissions,
        if (isActive != null) 'is_active': isActive,
      },
    );
    return StaffModel.fromJson(response.data['data'] as Map<String, dynamic>);
  }

  Future<String> resetPassword(String id) async {
    final response = await _client.dio.post(
      '${ApiConstants.users}/$id/reset-password',
    );
    return response.data['data']['temporary_password'] as String;
  }
}
