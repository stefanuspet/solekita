import 'package:mobile/core/constants/api_constants.dart';
import 'package:mobile/core/network/api_client.dart';
import 'package:mobile/domain/models/treatment_model.dart';

class TreatmentRemote {
  final ApiClient _client;

  TreatmentRemote(this._client);

  Future<List<TreatmentModel>> listTreatments({
    bool? isActive,
    String? material,
  }) async {
    final response = await _client.dio.get(
      ApiConstants.treatments,
      queryParameters: {
        if (isActive != null) 'is_active': isActive,
        if (material != null) 'material': material,
      },
    );
    return (response.data['data'] as List<dynamic>)
        .map((e) => TreatmentModel.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<TreatmentModel> createTreatment({
    required String name,
    required String material,
    required int price,
  }) async {
    final response = await _client.dio.post(
      ApiConstants.treatments,
      data: {'name': name, 'material': material, 'price': price},
    );
    return TreatmentModel.fromJson(
        response.data['data'] as Map<String, dynamic>);
  }

  Future<TreatmentModel> updateTreatment(
    String id, {
    String? name,
    int? price,
    bool? isActive,
  }) async {
    final response = await _client.dio.patch(
      '${ApiConstants.treatments}/$id',
      data: {
        if (name != null) 'name': name,
        if (price != null) 'price': price,
        if (isActive != null) 'is_active': isActive,
      },
    );
    return TreatmentModel.fromJson(
        response.data['data'] as Map<String, dynamic>);
  }

  Future<void> deleteTreatment(String id) async {
    await _client.dio.delete('${ApiConstants.treatments}/$id');
  }
}
