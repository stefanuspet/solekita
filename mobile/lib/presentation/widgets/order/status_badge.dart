import 'package:flutter/material.dart';
import 'package:mobile/core/theme/app_colors.dart';
import 'package:mobile/domain/models/order_model.dart';

/// Badge warna per status order. Gunakan [size] untuk mengatur ukuran teks.
class StatusBadge extends StatelessWidget {
  final OrderStatus status;
  final StatusBadgeSize size;

  const StatusBadge({
    super.key,
    required this.status,
    this.size = StatusBadgeSize.medium,
  });

  @override
  Widget build(BuildContext context) {
    final (bg, fg) = _palette(status);
    final (hPad, vPad, fontSize) = switch (size) {
      StatusBadgeSize.small => (8.0, 3.0, 10.0),
      StatusBadgeSize.medium => (10.0, 4.0, 11.0),
      StatusBadgeSize.large => (14.0, 6.0, 13.0),
    };

    return Container(
      padding: EdgeInsets.symmetric(horizontal: hPad, vertical: vPad),
      decoration: BoxDecoration(
        color: bg,
        borderRadius: BorderRadius.circular(20),
      ),
      child: Text(
        status.label,
        style: TextStyle(
          color: fg,
          fontSize: fontSize,
          fontWeight: FontWeight.w600,
          height: 1,
        ),
      ),
    );
  }

  static (Color bg, Color fg) _palette(OrderStatus s) => switch (s) {
        OrderStatus.dijemput =>
          (const Color(0xFFDBEAFE), const Color(0xFF1D4ED8)),
        OrderStatus.baru => (AppColors.surfaceTeal, AppColors.primaryDark),
        OrderStatus.proses =>
          (const Color(0xFFFEF3C7), const Color(0xFFB45309)),
        OrderStatus.selesai =>
          (const Color(0xFFDCFCE7), const Color(0xFF15803D)),
        OrderStatus.diambil =>
          (AppColors.surfaceAlt, AppColors.textSecondary),
        OrderStatus.diantar =>
          (AppColors.surfaceAlt, AppColors.textSecondary),
        OrderStatus.dibatalkan =>
          (const Color(0xFFFEE2E2), const Color(0xFFB91C1C)),
      };
}

enum StatusBadgeSize { small, medium, large }
