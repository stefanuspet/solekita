import 'dart:io';

import 'package:dio/dio.dart';
import 'package:mobile/core/constants/api_constants.dart';
import 'package:mobile/core/network/api_client.dart';
import 'package:mobile/domain/models/attendance_model.dart';

class AttendanceRemote {
  final ApiClient _client;

  AttendanceRemote(this._client);

  Future<AttendanceModel> checkIn({File? selfie}) async {
    final formData = FormData.fromMap({
      'type': 'masuk',
      if (selfie != null)
        'selfie': await MultipartFile.fromFile(selfie.path),
    });
    final response = await _client.dio.post(
      ApiConstants.attendances,
      data: formData,
    );
    return AttendanceModel.fromJson(
        response.data['data'] as Map<String, dynamic>);
  }

  Future<AttendanceModel> checkOut({File? selfie}) async {
    final formData = FormData.fromMap({
      'type': 'keluar',
      if (selfie != null)
        'selfie': await MultipartFile.fromFile(selfie.path),
    });
    final response = await _client.dio.post(
      ApiConstants.attendances,
      data: formData,
    );
    return AttendanceModel.fromJson(
        response.data['data'] as Map<String, dynamic>);
  }

  Future<AttendanceTodayStatus> getToday() async {
    final response = await _client.dio.get(ApiConstants.todayAttendances);
    return AttendanceTodayStatus.fromJson(
        response.data['data'] as Map<String, dynamic>);
  }
}
