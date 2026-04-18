class ApiException implements Exception {
  final int statusCode;
  final String message;
  final Map<String, String>? errors;

  ApiException({
    required this.statusCode,
    required this.message,
    this.errors,
  });

  factory ApiException.fromJson(int statusCode, Map<String, dynamic> json) {
    return ApiException(
      statusCode: statusCode,
      message: json['message'] ?? 'Terjadi kesalahan',
      errors: (json['errors'] as Map?)?.map(
        (key, value) => MapEntry(key.toString(), value.toString()),
      ),
    );
  }

  @override
  String toString() {
    return 'ApiException($statusCode): $message';
  }
}
