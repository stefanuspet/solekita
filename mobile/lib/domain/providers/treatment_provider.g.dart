// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'treatment_provider.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

String _$treatmentRemoteHash() => r'1d3767395ecfe919382935f5067b70e6312fc007';

/// See also [treatmentRemote].
@ProviderFor(treatmentRemote)
final treatmentRemoteProvider = AutoDisposeProvider<TreatmentRemote>.internal(
  treatmentRemote,
  name: r'treatmentRemoteProvider',
  debugGetCreateSourceHash: const bool.fromEnvironment('dart.vm.product')
      ? null
      : _$treatmentRemoteHash,
  dependencies: null,
  allTransitiveDependencies: null,
);

typedef TreatmentRemoteRef = AutoDisposeProviderRef<TreatmentRemote>;
String _$activeTreatmentsHash() => r'd24b9ad8d0bac8361bc72a7a40797561fca7b2bb';

/// Copied from Dart SDK
class _SystemHash {
  _SystemHash._();

  static int combine(int hash, int value) {
    // ignore: parameter_assignments
    hash = 0x1fffffff & (hash + value);
    // ignore: parameter_assignments
    hash = 0x1fffffff & (hash + ((0x0007ffff & hash) << 10));
    return hash ^ (hash >> 6);
  }

  static int finish(int hash) {
    // ignore: parameter_assignments
    hash = 0x1fffffff & (hash + ((0x03ffffff & hash) << 3));
    // ignore: parameter_assignments
    hash = hash ^ (hash >> 11);
    return 0x1fffffff & (hash + ((0x00003fff & hash) << 15));
  }
}

/// List treatment aktif — di-cache Riverpod, invalidate manual saat setting berubah.
/// Contoh invalidate: ref.invalidate(activeTreatmentsProvider)
///
/// Copied from [activeTreatments].
@ProviderFor(activeTreatments)
const activeTreatmentsProvider = ActiveTreatmentsFamily();

/// List treatment aktif — di-cache Riverpod, invalidate manual saat setting berubah.
/// Contoh invalidate: ref.invalidate(activeTreatmentsProvider)
///
/// Copied from [activeTreatments].
class ActiveTreatmentsFamily extends Family<AsyncValue<List<TreatmentModel>>> {
  /// List treatment aktif — di-cache Riverpod, invalidate manual saat setting berubah.
  /// Contoh invalidate: ref.invalidate(activeTreatmentsProvider)
  ///
  /// Copied from [activeTreatments].
  const ActiveTreatmentsFamily();

  /// List treatment aktif — di-cache Riverpod, invalidate manual saat setting berubah.
  /// Contoh invalidate: ref.invalidate(activeTreatmentsProvider)
  ///
  /// Copied from [activeTreatments].
  ActiveTreatmentsProvider call({
    String? material,
  }) {
    return ActiveTreatmentsProvider(
      material: material,
    );
  }

  @override
  ActiveTreatmentsProvider getProviderOverride(
    covariant ActiveTreatmentsProvider provider,
  ) {
    return call(
      material: provider.material,
    );
  }

  static const Iterable<ProviderOrFamily>? _dependencies = null;

  @override
  Iterable<ProviderOrFamily>? get dependencies => _dependencies;

  static const Iterable<ProviderOrFamily>? _allTransitiveDependencies = null;

  @override
  Iterable<ProviderOrFamily>? get allTransitiveDependencies =>
      _allTransitiveDependencies;

  @override
  String? get name => r'activeTreatmentsProvider';
}

/// List treatment aktif — di-cache Riverpod, invalidate manual saat setting berubah.
/// Contoh invalidate: ref.invalidate(activeTreatmentsProvider)
///
/// Copied from [activeTreatments].
class ActiveTreatmentsProvider
    extends AutoDisposeFutureProvider<List<TreatmentModel>> {
  /// List treatment aktif — di-cache Riverpod, invalidate manual saat setting berubah.
  /// Contoh invalidate: ref.invalidate(activeTreatmentsProvider)
  ///
  /// Copied from [activeTreatments].
  ActiveTreatmentsProvider({
    String? material,
  }) : this._internal(
          (ref) => activeTreatments(
            ref as ActiveTreatmentsRef,
            material: material,
          ),
          from: activeTreatmentsProvider,
          name: r'activeTreatmentsProvider',
          debugGetCreateSourceHash:
              const bool.fromEnvironment('dart.vm.product')
                  ? null
                  : _$activeTreatmentsHash,
          dependencies: ActiveTreatmentsFamily._dependencies,
          allTransitiveDependencies:
              ActiveTreatmentsFamily._allTransitiveDependencies,
          material: material,
        );

  ActiveTreatmentsProvider._internal(
    super._createNotifier, {
    required super.name,
    required super.dependencies,
    required super.allTransitiveDependencies,
    required super.debugGetCreateSourceHash,
    required super.from,
    required this.material,
  }) : super.internal();

  final String? material;

  @override
  Override overrideWith(
    FutureOr<List<TreatmentModel>> Function(ActiveTreatmentsRef provider)
        create,
  ) {
    return ProviderOverride(
      origin: this,
      override: ActiveTreatmentsProvider._internal(
        (ref) => create(ref as ActiveTreatmentsRef),
        from: from,
        name: null,
        dependencies: null,
        allTransitiveDependencies: null,
        debugGetCreateSourceHash: null,
        material: material,
      ),
    );
  }

  @override
  AutoDisposeFutureProviderElement<List<TreatmentModel>> createElement() {
    return _ActiveTreatmentsProviderElement(this);
  }

  @override
  bool operator ==(Object other) {
    return other is ActiveTreatmentsProvider && other.material == material;
  }

  @override
  int get hashCode {
    var hash = _SystemHash.combine(0, runtimeType.hashCode);
    hash = _SystemHash.combine(hash, material.hashCode);

    return _SystemHash.finish(hash);
  }
}

mixin ActiveTreatmentsRef
    on AutoDisposeFutureProviderRef<List<TreatmentModel>> {
  /// The parameter `material` of this provider.
  String? get material;
}

class _ActiveTreatmentsProviderElement
    extends AutoDisposeFutureProviderElement<List<TreatmentModel>>
    with ActiveTreatmentsRef {
  _ActiveTreatmentsProviderElement(super.provider);

  @override
  String? get material => (origin as ActiveTreatmentsProvider).material;
}
// ignore_for_file: type=lint
// ignore_for_file: subtype_of_sealed_class, invalid_use_of_internal_member, invalid_use_of_visible_for_testing_member
