// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'attendance_provider.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

String _$attendanceRemoteHash() => r'598b8c1f2711862a8933843d905039eaf70a3d3b';

/// See also [attendanceRemote].
@ProviderFor(attendanceRemote)
final attendanceRemoteProvider = AutoDisposeProvider<AttendanceRemote>.internal(
  attendanceRemote,
  name: r'attendanceRemoteProvider',
  debugGetCreateSourceHash: const bool.fromEnvironment('dart.vm.product')
      ? null
      : _$attendanceRemoteHash,
  dependencies: null,
  allTransitiveDependencies: null,
);

typedef AttendanceRemoteRef = AutoDisposeProviderRef<AttendanceRemote>;
String _$todayAttendanceHash() => r'29857b49c59bfe23b7ea4877dc66c650ecb49da4';

/// Status absensi hari ini. Invalidate setelah check-in/out berhasil.
///
/// Copied from [todayAttendance].
@ProviderFor(todayAttendance)
final todayAttendanceProvider =
    AutoDisposeFutureProvider<AttendanceTodayStatus>.internal(
  todayAttendance,
  name: r'todayAttendanceProvider',
  debugGetCreateSourceHash: const bool.fromEnvironment('dart.vm.product')
      ? null
      : _$todayAttendanceHash,
  dependencies: null,
  allTransitiveDependencies: null,
);

typedef TodayAttendanceRef
    = AutoDisposeFutureProviderRef<AttendanceTodayStatus>;
// ignore_for_file: type=lint
// ignore_for_file: subtype_of_sealed_class, invalid_use_of_internal_member, invalid_use_of_visible_for_testing_member
