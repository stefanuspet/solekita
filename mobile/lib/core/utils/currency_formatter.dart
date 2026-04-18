import 'package:intl/intl.dart';

class CurrencyFormatter {
  static final _formatter = NumberFormat.currency(
    locale: 'id_ID',
    symbol: 'Rp',
    decimalDigits: 0,
  );

  /// Format ke Rupiah: 35000 -> Rp35.000
  static String rupiah(num amount) {
    return _formatter.format(amount);
  }

  /// Format tanpa simbol: 35000 -> 35.000
  static String rupiahWithoutSymbol(num amount) {
    final formatter = NumberFormat.decimalPattern('id_ID');
    return formatter.format(amount);
  }

  /// Format dengan desimal: 35000.5 -> Rp35.000,50
  static String rupiahWithDecimal(num amount) {
    final formatter = NumberFormat.currency(
      locale: 'id_ID',
      symbol: 'Rp',
      decimalDigits: 2,
    );
    return formatter.format(amount);
  }

  /// Format compact: 1500000 -> Rp1,5 jt
  static String rupiahCompact(num amount) {
    if (amount >= 1000000) {
      return 'Rp${(amount / 1000000).toStringAsFixed(1)} jt';
    } else if (amount >= 1000) {
      return 'Rp${(amount / 1000).toStringAsFixed(1)} rb';
    } else {
      return 'Rp$amount';
    }
  }

  /// Parse dari string ke number: "35.000" -> 35000
  static int parse(String value) {
    final cleaned = value.replaceAll(RegExp(r'[^0-9]'), '');
    return int.tryParse(cleaned) ?? 0;
  }
}
