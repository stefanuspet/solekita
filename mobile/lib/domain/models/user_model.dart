class OutletInfo {
  final String id;
  final String name;
  final String code;

  const OutletInfo({
    required this.id,
    required this.name,
    required this.code,
  });

  factory OutletInfo.fromJson(Map<String, dynamic> json) => OutletInfo(
        id: json['id'] as String,
        name: json['name'] as String,
        code: json['code'] as String,
      );

  Map<String, dynamic> toJson() => {
        'id': id,
        'name': name,
        'code': code,
      };
}

class UserModel {
  final String id;
  final String name;
  final String phone;
  final bool isOwner;
  final OutletInfo outlet;
  final List<String> permissions;

  const UserModel({
    required this.id,
    required this.name,
    required this.phone,
    required this.isOwner,
    required this.outlet,
    required this.permissions,
  });

  factory UserModel.fromJson(Map<String, dynamic> json) => UserModel(
        id: json['id'] as String,
        name: json['name'] as String,
        phone: json['phone'] as String,
        isOwner: json['is_owner'] as bool,
        outlet: OutletInfo.fromJson(json['outlet'] as Map<String, dynamic>),
        permissions: (json['permissions'] as List<dynamic>)
            .map((e) => e as String)
            .toList(),
      );

  Map<String, dynamic> toJson() => {
        'id': id,
        'name': name,
        'phone': phone,
        'is_owner': isOwner,
        'outlet': outlet.toJson(),
        'permissions': permissions,
      };

  bool hasPermission(String permission) =>
      isOwner || permissions.contains(permission);
}
