import 'package:flutter/material.dart';
import 'package:mobile/core/theme/app_colors.dart';

class LoadingOverlay extends StatelessWidget {
  final bool isLoading;
  final Widget child;

  const LoadingOverlay({
    super.key,
    required this.isLoading,
    required this.child,
  });

  @override
  Widget build(BuildContext context) {
    return Stack(
      children: [
        child,
        if (isLoading)
          const Positioned.fill(
            child: _Overlay(),
          ),
      ],
    );
  }
}

class _Overlay extends StatelessWidget {
  const _Overlay();

  @override
  Widget build(BuildContext context) {
    return ColoredBox(
      color: Colors.black.withOpacity(0.35),
      child: const Center(
        child: _Spinner(),
      ),
    );
  }
}

class _Spinner extends StatelessWidget {
  const _Spinner();

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 72,
      height: 72,
      decoration: BoxDecoration(
        color: AppColors.surface,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.12),
            blurRadius: 16,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: const Center(
        child: CircularProgressIndicator(
          color: AppColors.primary,
          strokeWidth: 3,
        ),
      ),
    );
  }
}

/// Gunakan [showLoadingOverlay] untuk tampilkan overlay via route dialog
/// tanpa mengubah state widget parent.
Future<T> showLoadingOverlay<T>(
  BuildContext context,
  Future<T> Function() task,
) async {
  final overlay = OverlayEntry(builder: (_) => const _Overlay());
  Overlay.of(context).insert(overlay);
  try {
    return await task();
  } finally {
    overlay.remove();
  }
}
