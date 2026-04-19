class CustomerModel {
  final String id;
  final String name;
  final String phone;
  final int totalOrders;
  final DateTime? lastOrderAt;
  final bool isBlacklisted;
  final bool? isNew;

  const CustomerModel({
    required this.id,
    required this.name,
    required this.phone,
    required this.totalOrders,
    this.lastOrderAt,
    required this.isBlacklisted,
    this.isNew,
  });

  factory CustomerModel.fromJson(Map<String, dynamic> json) => CustomerModel(
        id: json['id'] as String,
        name: json['name'] as String,
        phone: json['phone'] as String,
        totalOrders: json['total_orders'] as int,
        lastOrderAt: json['last_order_at'] != null
            ? DateTime.parse(json['last_order_at'] as String)
            : null,
        isBlacklisted: json['is_blacklisted'] as bool,
        isNew: json['is_new'] as bool?,
      );

  Map<String, dynamic> toJson() => {
        'id': id,
        'name': name,
        'phone': phone,
        'total_orders': totalOrders,
        'last_order_at': lastOrderAt?.toIso8601String(),
        'is_blacklisted': isBlacklisted,
        'is_new': isNew,
      };
}
