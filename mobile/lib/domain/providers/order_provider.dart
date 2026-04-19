import 'dart:io';

import 'package:riverpod_annotation/riverpod_annotation.dart';
import 'package:mobile/data/remote/order_remote.dart';
import 'package:mobile/domain/models/order_model.dart';
import 'package:mobile/domain/providers/auth_provider.dart';

part 'order_provider.g.dart';

@riverpod
OrderRemote orderRemote(OrderRemoteRef ref) =>
    OrderRemote(ref.watch(apiClientProvider));

/// List order hari ini. Invalidate untuk refresh: ref.invalidate(todayOrdersProvider)
@riverpod
Future<OrderListResult> todayOrders(
  TodayOrdersRef ref, {
  String? status,
  String? search,
}) {
  final today = DateTime.now();
  final date =
      '${today.year}-${today.month.toString().padLeft(2, '0')}-${today.day.toString().padLeft(2, '0')}';
  return ref.watch(orderRemoteProvider).listOrders(
        date: date,
        status: status,
        search: search,
      );
}

/// Detail order by ID. Family provider — satu instance per orderId.
@riverpod
Future<OrderModel> orderDetail(OrderDetailRef ref, String orderId) =>
    ref.watch(orderRemoteProvider).getOrder(orderId);

// ── Create Order state machine ───────────────────────────────────────────────
// State: null = idle, AsyncLoading = loading, AsyncData(order) = success,
//        AsyncError = error

@riverpod
class CreateOrderNotifier extends _$CreateOrderNotifier {
  @override
  AsyncValue<OrderModel?> build() => const AsyncData(null);

  Future<void> createOrder({
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
    state = const AsyncLoading();
    state = await AsyncValue.guard(() => ref.read(orderRemoteProvider).createOrder(
          customerId: customerId,
          treatmentId: treatmentId,
          basePrice: basePrice,
          deliveryFee: deliveryFee,
          totalPrice: totalPrice,
          paymentMethod: paymentMethod,
          photoBefore: photoBefore,
          photoAfter: photoAfter,
          conditionNotes: conditionNotes,
          isPickup: isPickup,
          isDelivery: isDelivery,
          pickupAddress: pickupAddress,
          deliveryAddress: deliveryAddress,
        ));

    // Setelah order berhasil dibuat, refresh list order hari ini
    if (state is AsyncData) {
      ref.invalidate(todayOrdersProvider);
    }
  }

  void reset() => state = const AsyncData(null);
}
