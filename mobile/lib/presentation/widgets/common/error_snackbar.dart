import 'package:flutter/material.dart';
import 'package:mobile/core/theme/app_colors.dart';

/// Tampilkan snackbar error.
/// Contoh: ErrorSnackbar.show(context, 'Gagal menyimpan data');
class ErrorSnackbar {
  ErrorSnackbar._();

  static void show(BuildContext context, String message) {
    ScaffoldMessenger.of(context)
      ..hideCurrentSnackBar()
      ..showSnackBar(_buildSnackBar(message, AppColors.error, Icons.error_outline));
  }

  static void showSuccess(BuildContext context, String message) {
    ScaffoldMessenger.of(context)
      ..hideCurrentSnackBar()
      ..showSnackBar(_buildSnackBar(message, AppColors.success, Icons.check_circle_outline));
  }

  static void showInfo(BuildContext context, String message) {
    ScaffoldMessenger.of(context)
      ..hideCurrentSnackBar()
      ..showSnackBar(_buildSnackBar(message, AppColors.info, Icons.info_outline));
  }

  static SnackBar _buildSnackBar(String message, Color color, IconData icon) {
    return SnackBar(
      behavior: SnackBarBehavior.floating,
      margin: const EdgeInsets.fromLTRB(16, 0, 16, 16),
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      backgroundColor: AppColors.textPrimary,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      duration: const Duration(seconds: 4),
      content: Row(
        children: [
          Icon(icon, color: color, size: 20),
          const SizedBox(width: 10),
          Expanded(
            child: Text(
              message,
              style: const TextStyle(
                color: Colors.white,
                fontSize: 14,
                fontWeight: FontWeight.w500,
              ),
            ),
          ),
        ],
      ),
    );
  }
}
