import 'package:riverpod_annotation/riverpod_annotation.dart';
import 'package:mobile/data/remote/treatment_remote.dart';
import 'package:mobile/domain/models/treatment_model.dart';
import 'package:mobile/domain/providers/auth_provider.dart';

part 'treatment_provider.g.dart';

@riverpod
TreatmentRemote treatmentRemote(TreatmentRemoteRef ref) =>
    TreatmentRemote(ref.watch(apiClientProvider));

/// List treatment aktif — di-cache Riverpod, invalidate manual saat setting berubah.
/// Contoh invalidate: ref.invalidate(activeTreatmentsProvider)
@riverpod
Future<List<TreatmentModel>> activeTreatments(
  ActiveTreatmentsRef ref, {
  String? material,
}) =>
    ref.watch(treatmentRemoteProvider).listTreatments(
          isActive: true,
          material: material,
        );
