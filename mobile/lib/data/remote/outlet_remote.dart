import 'package:mobile/core/constants/api_constants.dart';
import 'package:mobile/core/network/api_client.dart';
import 'package:mobile/domain/models/outlet_model.dart';

class OverdueOrder {
  final String id;
  final String orderNumber;
  final String customerName;
  final String customerPhone;
  final String status;
  final int totalPrice;
  final int daysOverdue;

  const OverdueOrder({
    required this.id,
    required this.orderNumber,
    required this.customerName,
    required this.customerPhone,
    required this.status,
    required this.totalPrice,
    required this.daysOverdue,
  });

  factory OverdueOrder.fromJson(Map<String, dynamic> json) => OverdueOrder(
        id: json['id'] as String,
        orderNumber: json['order_number'] as String,
        customerName: json['customer_name'] as String,
        customerPhone: json['customer_phone'] as String,
        status: json['status'] as String,
        totalPrice: json['total_price'] as int,
        daysOverdue: json['days_overdue'] as int,
      );
}

class StaffSummary {
  final String userId;
  final String name;
  final int orderCount;
  final int totalRevenue;

  const StaffSummary({
    required this.userId,
    required this.name,
    required this.orderCount,
    required this.totalRevenue,
  });

  factory StaffSummary.fromJson(Map<String, dynamic> json) => StaffSummary(
        userId: json['user_id'] as String,
        name: json['name'] as String,
        orderCount: json['order_count'] as int,
        totalRevenue: json['total_revenue'] as int,
      );
}

class OutletSummary {
  final int todayRevenue;
  final int todayOrderCount;
  final int todayPaidCash;
  final int todayPaidTransfer;
  final int todayPaidQris;
  final int overdueCount;
  final List<OverdueOrder> overdueOrders;
  final List<StaffSummary> staffSummary;

  const OutletSummary({
    required this.todayRevenue,
    required this.todayOrderCount,
    required this.todayPaidCash,
    required this.todayPaidTransfer,
    required this.todayPaidQris,
    required this.overdueCount,
    required this.overdueOrders,
    required this.staffSummary,
  });

  factory OutletSummary.fromJson(Map<String, dynamic> json) {
    final today = json['today'] as Map<String, dynamic>;
    final overdue = json['overdue_orders'] as Map<String, dynamic>;
    return OutletSummary(
      todayRevenue: today['revenue'] as int,
      todayOrderCount: today['order_count'] as int,
      todayPaidCash: today['paid_cash'] as int,
      todayPaidTransfer: today['paid_transfer'] as int,
      todayPaidQris: today['paid_qris'] as int,
      overdueCount: overdue['count'] as int,
      overdueOrders: (overdue['orders'] as List<dynamic>)
          .map((e) => OverdueOrder.fromJson(e as Map<String, dynamic>))
          .toList(),
      staffSummary: (json['staff_summary'] as List<dynamic>)
          .map((e) => StaffSummary.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }
}

class OutletRemote {
  final ApiClient _client;

  OutletRemote(this._client);

  Future<OutletModel> getMyOutlet() async {
    final response = await _client.dio.get(ApiConstants.outletMe);
    return OutletModel.fromJson(response.data['data'] as Map<String, dynamic>);
  }

  Future<OutletModel> updateMyOutlet({
    String? name,
    String? address,
    String? phone,
    int? overdueThresholdDays,
  }) async {
    final response = await _client.dio.patch(
      ApiConstants.outletMe,
      data: {
        if (name != null) 'name': name,
        if (address != null) 'address': address,
        if (phone != null) 'phone': phone,
        if (overdueThresholdDays != null)
          'overdue_threshold_days': overdueThresholdDays,
      },
    );
    return OutletModel.fromJson(response.data['data'] as Map<String, dynamic>);
  }

  Future<OutletSummary> getSummary() async {
    final response = await _client.dio.get(ApiConstants.outletSummary);
    return OutletSummary.fromJson(
        response.data['data'] as Map<String, dynamic>);
  }
}
