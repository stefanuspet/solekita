import 'package:mobile/core/constants/api_constants.dart';
import 'package:mobile/core/network/api_client.dart';
import 'package:mobile/domain/models/customer_model.dart';

typedef FindOrCreateResult = ({CustomerModel customer, bool isNew});

class CustomerRemote {
  final ApiClient _client;

  CustomerRemote(this._client);

  Future<FindOrCreateResult> findOrCreate({
    required String name,
    required String phone,
  }) async {
    final response = await _client.dio.post(
      ApiConstants.customers,
      data: {'name': name, 'phone': phone},
    );
    final data = response.data['data'] as Map<String, dynamic>;
    return (
      customer: CustomerModel.fromJson(data),
      isNew: data['is_new'] as bool? ?? false,
    );
  }

  Future<List<CustomerModel>> searchCustomers(String query) async {
    final response = await _client.dio.get(
      ApiConstants.customers,
      queryParameters: {'search': query, 'limit': 20},
    );
    final items = response.data['data']['items'] as List<dynamic>;
    return items
        .map((e) => CustomerModel.fromJson(e as Map<String, dynamic>))
        .toList();
  }
}
