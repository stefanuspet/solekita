import 'package:riverpod_annotation/riverpod_annotation.dart';
import 'package:mobile/data/remote/attendance_remote.dart';
import 'package:mobile/domain/models/attendance_model.dart';
import 'package:mobile/domain/providers/auth_provider.dart';

part 'attendance_provider.g.dart';

@riverpod
AttendanceRemote attendanceRemote(AttendanceRemoteRef ref) =>
    AttendanceRemote(ref.watch(apiClientProvider));

/// Status absensi hari ini. Invalidate setelah check-in/out berhasil.
@riverpod
Future<AttendanceTodayStatus> todayAttendance(TodayAttendanceRef ref) =>
    ref.watch(attendanceRemoteProvider).getToday();
