class OutletModel {
  final String id;
  final String name;
  final String code;
  final String address;
  final String phone;
  final String subscriptionStatus;
  final int overdueThresholdDays;
  final bool isActive;
  final DateTime createdAt;

  const OutletModel({
    required this.id,
    required this.name,
    required this.code,
    required this.address,
    required this.phone,
    required this.subscriptionStatus,
    required this.overdueThresholdDays,
    required this.isActive,
    required this.createdAt,
  });

  factory OutletModel.fromJson(Map<String, dynamic> json) => OutletModel(
        id: json['id'] as String,
        name: json['name'] as String,
        code: json['code'] as String,
        address: json['address'] as String,
        phone: json['phone'] as String,
        subscriptionStatus: json['subscription_status'] as String,
        overdueThresholdDays: json['overdue_threshold_days'] as int,
        isActive: json['is_active'] as bool,
        createdAt: DateTime.parse(json['created_at'] as String),
      );

  Map<String, dynamic> toJson() => {
        'id': id,
        'name': name,
        'code': code,
        'address': address,
        'phone': phone,
        'subscription_status': subscriptionStatus,
        'overdue_threshold_days': overdueThresholdDays,
        'is_active': isActive,
        'created_at': createdAt.toIso8601String(),
      };

  bool get isTrial => subscriptionStatus == 'trial';
  bool get isSubscriptionActive => subscriptionStatus == 'active';
  bool get isSuspended => subscriptionStatus == 'suspended';
}
