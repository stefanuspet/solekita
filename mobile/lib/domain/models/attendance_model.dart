class AttendanceUser {
  final String id;
  final String name;

  const AttendanceUser({required this.id, required this.name});

  factory AttendanceUser.fromJson(Map<String, dynamic> json) => AttendanceUser(
        id: json['id'] as String,
        name: json['name'] as String,
      );

  Map<String, dynamic> toJson() => {'id': id, 'name': name};
}

class AttendanceModel {
  final String id;
  final AttendanceUser? user;
  final String type; // "masuk" | "keluar"
  final String? selfieUrl;
  final DateTime createdAt;

  const AttendanceModel({
    required this.id,
    this.user,
    required this.type,
    this.selfieUrl,
    required this.createdAt,
  });

  factory AttendanceModel.fromJson(Map<String, dynamic> json) =>
      AttendanceModel(
        id: json['id'] as String,
        user: json['user'] != null
            ? AttendanceUser.fromJson(json['user'] as Map<String, dynamic>)
            : null,
        type: json['type'] as String,
        selfieUrl: json['selfie_url'] as String?,
        createdAt: DateTime.parse(json['created_at'] as String),
      );

  Map<String, dynamic> toJson() => {
        'id': id,
        'user': user?.toJson(),
        'type': type,
        'selfie_url': selfieUrl,
        'created_at': createdAt.toIso8601String(),
      };
}

class AttendanceTodayStatus {
  final bool hasCheckedIn;
  final bool hasCheckedOut;
  final DateTime? checkedInAt;
  final DateTime? checkedOutAt;

  const AttendanceTodayStatus({
    required this.hasCheckedIn,
    required this.hasCheckedOut,
    this.checkedInAt,
    this.checkedOutAt,
  });

  factory AttendanceTodayStatus.fromJson(Map<String, dynamic> json) =>
      AttendanceTodayStatus(
        hasCheckedIn: json['has_checked_in'] as bool,
        hasCheckedOut: json['has_checked_out'] as bool,
        checkedInAt: json['checked_in_at'] != null
            ? DateTime.parse(json['checked_in_at'] as String)
            : null,
        checkedOutAt: json['checked_out_at'] != null
            ? DateTime.parse(json['checked_out_at'] as String)
            : null,
      );

  Map<String, dynamic> toJson() => {
        'has_checked_in': hasCheckedIn,
        'has_checked_out': hasCheckedOut,
        'checked_in_at': checkedInAt?.toIso8601String(),
        'checked_out_at': checkedOutAt?.toIso8601String(),
      };
}
