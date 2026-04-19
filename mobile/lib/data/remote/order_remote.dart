import 'dart:io';

import 'package:dio/dio.dart';
import 'package:mobile/core/constants/api_constants.dart';
import 'package:mobile/core/network/api_client.dart';
import 'package:mobile/domain/models/order_model.dart';

class OrderListItem {
  final String id;
  final String orderNumber;
  final OrderCustomer customer;
  final OrderKasir kasir;
  final String treatmentName;
  final String material;
  final OrderStatus status;
  final int totalPrice;
  final bool isPickup;
  final bool isDelivery;
  final bool hasPhotoBefore;
  final bool hasPhotoAfter;
  final DateTime createdAt;
  final DateTime updatedAt;

  const OrderListItem({
    required this.id,
    required this.orderNumber,
    required this.customer,
    required this.kasir,
    required this.treatmentName,
    required this.material,
    required this.status,
    required this.totalPrice,
    required this.isPickup,
    required this.isDelivery,
    required this.hasPhotoBefore,
    required this.hasPhotoAfter,
    required this.createdAt,
    required this.updatedAt,
  });

  factory OrderListItem.fromJson(Map<String, dynamic> json) => OrderListItem(
        id: json['id'] as String,
        orderNumber: json['order_number'] as String,
        customer:
            OrderCustomer.fromJson(json['customer'] as Map<String, dynamic>),
        kasir: OrderKasir.fromJson(json['kasir'] as Map<String, dynamic>),
        treatmentName: json['treatment_name'] as String,
        material: json['material'] as String,
        status: OrderStatus.fromString(json['status'] as String),
        totalPrice: json['total_price'] as int,
        isPickup: json['is_pickup'] as bool,
        isDelivery: json['is_delivery'] as bool,
        hasPhotoBefore: json['has_photo_before'] as bool,
        hasPhotoAfter: json['has_photo_after'] as bool,
        createdAt: DateTime.parse(json['created_at'] as String),
        updatedAt: DateTime.parse(json['updated_at'] as String),
      );
}

class PaginationMeta {
  final int page;
  final int limit;
  final int total;
  final int totalPages;

  const PaginationMeta({
    required this.page,
    required this.limit,
    required this.total,
    required this.totalPages,
  });

  factory PaginationMeta.fromJson(Map<String, dynamic> json) => PaginationMeta(
        page: json['page'] as int,
        limit: json['limit'] as int,
        total: json['total'] as int,
        totalPages: json['total_pages'] as int,
      );
}

class OrderListResult {
  final List<OrderListItem> items;
  final PaginationMeta pagination;

  const OrderListResult({required this.items, required this.pagination});
}

class OrderRemote {
  final ApiClient _client;

  OrderRemote(this._client);

  Future<OrderModel> createOrder({
    required String customerId,
    required String treatmentId,
    required int basePrice,
    required int deliveryFee,
    required int totalPrice,
    required String paymentMethod,
    required File photoBefore,
    File? photoAfter,
    String? conditionNotes,
    bool isPickup = false,
    bool isDelivery = false,
    String? pickupAddress,
    String? deliveryAddress,
  }) async {
    final formData = FormData.fromMap({
      'customer_id': customerId,
      'treatment_id': treatmentId,
      'base_price': basePrice,
      'delivery_fee': deliveryFee,
      'total_price': totalPrice,
      'payment_method': paymentMethod,
      'is_pickup': isPickup,
      'is_delivery': isDelivery,
      if (conditionNotes != null) 'condition_notes': conditionNotes,
      if (pickupAddress != null) 'pickup_address': pickupAddress,
      if (deliveryAddress != null) 'delivery_address': deliveryAddress,
      'photo_before': await MultipartFile.fromFile(photoBefore.path),
      if (photoAfter != null)
        'photo_after': await MultipartFile.fromFile(photoAfter.path),
    });

    final response = await _client.dio.post(
      ApiConstants.orders,
      data: formData,
    );
    return OrderModel.fromJson(response.data['data'] as Map<String, dynamic>);
  }

  Future<OrderListResult> listOrders({
    String? status,
    String? date,
    String? search,
    int page = 1,
    int limit = 20,
  }) async {
    final response = await _client.dio.get(
      ApiConstants.orders,
      queryParameters: {
        if (status != null) 'status': status,
        if (date != null) 'date': date,
        if (search != null) 'search': search,
        'page': page,
        'limit': limit,
      },
    );

    final data = response.data['data'] as Map<String, dynamic>;
    return OrderListResult(
      items: (data['items'] as List<dynamic>)
          .map((e) => OrderListItem.fromJson(e as Map<String, dynamic>))
          .toList(),
      pagination: PaginationMeta.fromJson(
          data['pagination'] as Map<String, dynamic>),
    );
  }

  Future<OrderModel> getOrder(String id) async {
    final response = await _client.dio.get('${ApiConstants.orders}/$id');
    return OrderModel.fromJson(response.data['data'] as Map<String, dynamic>);
  }

  Future<OrderModel> updateStatus(String id, OrderStatus status) async {
    await _client.dio.patch(
      '${ApiConstants.orders}/$id/status',
      data: {'status': status.name},
    );
    // Response PATCH hanya return partial — fetch full order setelah update
    return getOrder(id);
  }

  Future<void> cancelOrder(String id, String reason) async {
    await _client.dio.post(
      '${ApiConstants.orders}/$id/cancel',
      data: {'reason': reason},
    );
  }

  Future<void> editPrice(String id, int totalPrice, String reason) async {
    await _client.dio.patch(
      '${ApiConstants.orders}/$id/price',
      data: {'total_price': totalPrice, 'reason': reason},
    );
  }

  Future<void> uploadAfterPhoto(String id, File photo) async {
    final formData = FormData.fromMap({
      'type': 'after',
      'photo': await MultipartFile.fromFile(photo.path),
    });
    await _client.dio.post(
      '${ApiConstants.orders}/$id/photos',
      data: formData,
    );
  }
}
