import 'package:intl/intl.dart';

class DateFormatter {
  static final _dateFormatter = DateFormat('d MMM yyyy', 'id_ID');
  static final _dateTimeFormatter = DateFormat('d MMM yyyy, HH:mm', 'id_ID');
  static final _timeFormatter = DateFormat('HH:mm', 'id_ID');

  /// 11 Apr 2026
  static String date(DateTime date) {
    return _dateFormatter.format(date);
  }

  /// 11 Apr 2026, 09:00
  static String dateTime(DateTime date) {
    return _dateTimeFormatter.format(date);
  }

  /// 09:00
  static String time(DateTime date) {
    return _timeFormatter.format(date);
  }

  /// Hari ini / Kemarin / tanggal
  static String relative(DateTime date) {
    final now = DateTime.now();

    final today = DateTime(now.year, now.month, now.day);
    final target = DateTime(date.year, date.month, date.day);

    final difference = today.difference(target).inDays;

    if (difference == 0) {
      return 'Hari ini';
    } else if (difference == 1) {
      return 'Kemarin';
    } else {
      return _dateFormatter.format(date);
    }
  }
}
