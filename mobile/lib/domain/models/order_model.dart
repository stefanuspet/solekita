enum OrderStatus {
  dijemput,
  baru,
  proses,
  selesai,
  diambil,
  diantar,
  dibatalkan;

  static OrderStatus fromString(String value) {
    return OrderStatus.values.firstWhere(
      (e) => e.name == value,
      orElse: () => throw ArgumentError('Unknown OrderStatus: $value'),
    );
  }

  String get label => switch (this) {
        OrderStatus.dijemput => 'Dijemput',
        OrderStatus.baru => 'Baru',
        OrderStatus.proses => 'Proses',
        OrderStatus.selesai => 'Selesai',
        OrderStatus.diambil => 'Diambil',
        OrderStatus.diantar => 'Diantar',
        OrderStatus.dibatalkan => 'Dibatalkan',
      };

  // Status berikutnya yang valid
  List<OrderStatus> get nextStatuses => switch (this) {
        OrderStatus.dijemput => [OrderStatus.baru],
        OrderStatus.baru => [OrderStatus.proses],
        OrderStatus.proses => [OrderStatus.selesai],
        OrderStatus.selesai => [OrderStatus.diambil, OrderStatus.diantar],
        _ => [],
      };

  bool get isFinal =>
      this == OrderStatus.diambil ||
      this == OrderStatus.diantar ||
      this == OrderStatus.dibatalkan;

  bool get canBeCancelled =>
      this == OrderStatus.dijemput ||
      this == OrderStatus.baru ||
      this == OrderStatus.proses;
}

class OrderCustomer {
  final String id;
  final String name;
  final String phone;

  const OrderCustomer({
    required this.id,
    required this.name,
    required this.phone,
  });

  factory OrderCustomer.fromJson(Map<String, dynamic> json) => OrderCustomer(
        id: json['id'] as String,
        name: json['name'] as String,
        phone: json['phone'] as String,
      );

  Map<String, dynamic> toJson() => {'id': id, 'name': name, 'phone': phone};
}

class OrderKasir {
  final String id;
  final String name;

  const OrderKasir({required this.id, required this.name});

  factory OrderKasir.fromJson(Map<String, dynamic> json) => OrderKasir(
        id: json['id'] as String,
        name: json['name'] as String,
      );

  Map<String, dynamic> toJson() => {'id': id, 'name': name};
}

class OrderPhoto {
  final String? beforeUrl;
  final String? afterUrl;

  const OrderPhoto({this.beforeUrl, this.afterUrl});

  factory OrderPhoto.fromJson(Map<String, dynamic> json) => OrderPhoto(
        beforeUrl: json['before_url'] as String?,
        afterUrl: json['after_url'] as String?,
      );

  Map<String, dynamic> toJson() => {
        'before_url': beforeUrl,
        'after_url': afterUrl,
      };
}

class OrderPayment {
  final String id;
  final String method;
  final int amount;
  final String status;
  final DateTime? paidAt;

  const OrderPayment({
    required this.id,
    required this.method,
    required this.amount,
    required this.status,
    this.paidAt,
  });

  factory OrderPayment.fromJson(Map<String, dynamic> json) => OrderPayment(
        id: json['id'] as String,
        method: json['method'] as String,
        amount: json['amount'] as int,
        status: json['status'] as String,
        paidAt: json['paid_at'] != null
            ? DateTime.parse(json['paid_at'] as String)
            : null,
      );

  Map<String, dynamic> toJson() => {
        'id': id,
        'method': method,
        'amount': amount,
        'status': status,
        'paid_at': paidAt?.toIso8601String(),
      };
}

class OrderLog {
  final String action;
  final String userName;
  final String? oldValue;
  final String? newValue;
  final DateTime createdAt;

  const OrderLog({
    required this.action,
    required this.userName,
    this.oldValue,
    this.newValue,
    required this.createdAt,
  });

  factory OrderLog.fromJson(Map<String, dynamic> json) => OrderLog(
        action: json['action'] as String,
        userName: json['user_name'] as String,
        oldValue: json['old_value'] as String?,
        newValue: json['new_value'] as String?,
        createdAt: DateTime.parse(json['created_at'] as String),
      );

  Map<String, dynamic> toJson() => {
        'action': action,
        'user_name': userName,
        'old_value': oldValue,
        'new_value': newValue,
        'created_at': createdAt.toIso8601String(),
      };
}

class OrderModel {
  final String id;
  final String orderNumber;
  final OrderCustomer customer;
  final OrderKasir kasir;
  final String treatmentName;
  final String material;
  final OrderStatus status;
  final int basePrice;
  final int deliveryFee;
  final int totalPrice;
  final bool isPriceEdited;
  final int? originalPrice;
  final String? conditionNotes;
  final bool isPickup;
  final bool isDelivery;
  final OrderPhoto photos;
  final OrderPayment? payment;
  final String? cancelReason;
  final DateTime? cancelledAt;
  final List<OrderLog> logs;
  final DateTime createdAt;
  final DateTime updatedAt;

  const OrderModel({
    required this.id,
    required this.orderNumber,
    required this.customer,
    required this.kasir,
    required this.treatmentName,
    required this.material,
    required this.status,
    required this.basePrice,
    required this.deliveryFee,
    required this.totalPrice,
    required this.isPriceEdited,
    this.originalPrice,
    this.conditionNotes,
    required this.isPickup,
    required this.isDelivery,
    required this.photos,
    this.payment,
    this.cancelReason,
    this.cancelledAt,
    required this.logs,
    required this.createdAt,
    required this.updatedAt,
  });

  factory OrderModel.fromJson(Map<String, dynamic> json) => OrderModel(
        id: json['id'] as String,
        orderNumber: json['order_number'] as String,
        customer: OrderCustomer.fromJson(
            json['customer'] as Map<String, dynamic>),
        kasir: OrderKasir.fromJson(json['kasir'] as Map<String, dynamic>),
        treatmentName: json['treatment_name'] as String,
        material: json['material'] as String,
        status: OrderStatus.fromString(json['status'] as String),
        basePrice: json['base_price'] as int,
        deliveryFee: json['delivery_fee'] as int,
        totalPrice: json['total_price'] as int,
        isPriceEdited: json['is_price_edited'] as bool,
        originalPrice: json['original_price'] as int?,
        conditionNotes: json['condition_notes'] as String?,
        isPickup: json['is_pickup'] as bool,
        isDelivery: json['is_delivery'] as bool,
        photos: OrderPhoto.fromJson(json['photos'] as Map<String, dynamic>),
        payment: json['payment'] != null
            ? OrderPayment.fromJson(json['payment'] as Map<String, dynamic>)
            : null,
        cancelReason: json['cancel_reason'] as String?,
        cancelledAt: json['cancelled_at'] != null
            ? DateTime.parse(json['cancelled_at'] as String)
            : null,
        logs: (json['logs'] as List<dynamic>)
            .map((e) => OrderLog.fromJson(e as Map<String, dynamic>))
            .toList(),
        createdAt: DateTime.parse(json['created_at'] as String),
        updatedAt: DateTime.parse(json['updated_at'] as String),
      );

  Map<String, dynamic> toJson() => {
        'id': id,
        'order_number': orderNumber,
        'customer': customer.toJson(),
        'kasir': kasir.toJson(),
        'treatment_name': treatmentName,
        'material': material,
        'status': status.name,
        'base_price': basePrice,
        'delivery_fee': deliveryFee,
        'total_price': totalPrice,
        'is_price_edited': isPriceEdited,
        'original_price': originalPrice,
        'condition_notes': conditionNotes,
        'is_pickup': isPickup,
        'is_delivery': isDelivery,
        'photos': photos.toJson(),
        'payment': payment?.toJson(),
        'cancel_reason': cancelReason,
        'cancelled_at': cancelledAt?.toIso8601String(),
        'logs': logs.map((e) => e.toJson()).toList(),
        'created_at': createdAt.toIso8601String(),
        'updated_at': updatedAt.toIso8601String(),
      };
}
