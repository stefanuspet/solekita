class TreatmentModel {
  final String id;
  final String name;
  final String material;
  final int price;
  final bool isActive;
  final DateTime? createdAt;

  const TreatmentModel({
    required this.id,
    required this.name,
    required this.material,
    required this.price,
    required this.isActive,
    this.createdAt,
  });

  factory TreatmentModel.fromJson(Map<String, dynamic> json) => TreatmentModel(
        id: json['id'] as String,
        name: json['name'] as String,
        material: json['material'] as String,
        price: json['price'] as int,
        isActive: json['is_active'] as bool,
        createdAt: json['created_at'] != null
            ? DateTime.parse(json['created_at'] as String)
            : null,
      );

  Map<String, dynamic> toJson() => {
        'id': id,
        'name': name,
        'material': material,
        'price': price,
        'is_active': isActive,
        'created_at': createdAt?.toIso8601String(),
      };
}
