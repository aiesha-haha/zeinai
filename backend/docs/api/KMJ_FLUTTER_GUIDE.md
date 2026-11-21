# 📱 KMJ API - Flutter Integration Guide

## 🔗 API Endpoints Overview

All endpoints are prefixed with `/api/v1/kmj`

### Authentication
All requests require Bearer token:
```dart
final dio = Dio(BaseOptions(
  baseUrl: 'http://localhost:8000/api/v1',
  headers: {'Authorization': 'Bearer $accessToken'},
));
```

---

## 📋 1. Create Draft Plan

**Endpoint:** `POST /api/v1/kmj`

**Request:**
```dart
final response = await dio.post('/kmj', data: {
  'subject': 'Информатика',
  'grade': 8,
  'lesson_number': 16,
  'lesson_topic': 'Кестеде көрсетілген функция графигінің қасиеттерін өзгерту',
  'teacher_name': 'Арай Кенжебекова',
  'content_lang': 'kaz', // kaz, rus, or eng
  'lesson_date': '2024-05-15T00:00:00Z', // Optional
});
```

**Response (201 Created):**
```dart
{
  "success": true,
  "data": {
    "plan": {
      "id": "uuid-here",
      "owner_id": "uuid",
      "subject": "Информатика",
      "grade": 8,
      "lesson_number": 16,
      "lesson_topic": "...",
      "teacher_name": "Арай Кенжебекова",
      "content_lang": "kaz",
      "status": "draft", // статус плана
      "created_at": "2024-05-15T12:00:00Z",
      "updated_at": "2024-05-15T12:00:00Z"
    },
    "cells": [
      {
        "id": "cell-uuid",
        "plan_id": "plan-uuid",
        "section": "header", // header, lesson_start, lesson_middle, lesson_end
        "cell_key": "bolem",
        "cell_value": "", // пустое для черновика
        "stage_column": null,
        "is_ai_generated": false,
        "is_edited": false,
        "created_at": "...",
        "updated_at": "..."
      },
      // ... остальные ячейки
    ]
  }
}
```

**Dart Model:**
```dart
class KMJPlan {
  final String id;
  final String ownerId;
  final String? schoolId;
  final String subject;
  final int grade;
  final int lessonNumber;
  final String lessonTopic;
  final String teacherName;
  final DateTime? lessonDate;
  final String contentLang;
  final String status; // draft, generating, completed, failed
  final String? generationTaskId;
  final DateTime createdAt;
  final DateTime updatedAt;
  
  factory KMJPlan.fromJson(Map<String, dynamic> json) => KMJPlan(
    id: json['id'],
    ownerId: json['owner_id'],
    schoolId: json['school_id'],
    subject: json['subject'],
    grade: json['grade'],
    lessonNumber: json['lesson_number'],
    lessonTopic: json['lesson_topic'],
    teacherName: json['teacher_name'],
    lessonDate: json['lesson_date'] != null 
        ? DateTime.parse(json['lesson_date']) 
        : null,
    contentLang: json['content_lang'],
    status: json['status'],
    generationTaskId: json['generation_task_id'],
    createdAt: DateTime.parse(json['created_at']),
    updatedAt: DateTime.parse(json['updated_at']),
  );
}

class KMJCell {
  final String id;
  final String planId;
  final String section; // header, lesson_start, lesson_middle, lesson_end
  final String cellKey;
  final String cellValue;
  final String? stageColumn; // pedagog_action, student_action, assessment, resources
  final bool isAiGenerated;
  final bool isEdited;
  final DateTime createdAt;
  final DateTime updatedAt;
  
  factory KMJCell.fromJson(Map<String, dynamic> json) => KMJCell(
    id: json['id'],
    planId: json['plan_id'],
    section: json['section'],
    cellKey: json['cell_key'],
    cellValue: json['cell_value'],
    stageColumn: json['stage_column'],
    isAiGenerated: json['is_ai_generated'],
    isEdited: json['is_edited'],
    createdAt: DateTime.parse(json['created_at']),
    updatedAt: DateTime.parse(json['updated_at']),
  );
}

class KMJPlanWithCells {
  final KMJPlan plan;
  final List<KMJCell> cells;
  
  factory KMJPlanWithCells.fromJson(Map<String, dynamic> json) => KMJPlanWithCells(
    plan: KMJPlan.fromJson(json['plan']),
    cells: (json['cells'] as List).map((e) => KMJCell.fromJson(e)).toList(),
  );
}
```

---

## 🤖 2. Generate Plan with AI

**⚠️ ВАЖНО: Стоимость 5 токенов**

**Endpoint:** `POST /api/v1/kmj/{id}/generate`

**Request:**
```dart
final planId = 'your-plan-uuid';
try {
  final response = await dio.post('/kmj/$planId/generate');
  
  if (response.data['success']) {
    final message = response.data['data']['message'];
    print(message); // "Generation started successfully"
    
    // Начать polling для проверки статуса
    _pollPlanStatus(planId);
  }
} on DioException catch (e) {
  if (e.response?.statusCode == 403) {
    // Недостаточно токенов или не владелец
    final error = e.response?.data['error'];
    if (error['code'] == 'INSUFFICIENT_TOKENS') {
      showDialog('Недостаточно токенов. Требуется 5 токенов.');
    }
  }
}
```

**Polling статуса:**
```dart
Timer? _pollTimer;

void _pollPlanStatus(String planId) {
  _pollTimer = Timer.periodic(Duration(seconds: 3), (timer) async {
    try {
      final response = await dio.get('/kmj/$planId');
      final plan = KMJPlan.fromJson(response.data['data']['plan']);
      
      if (plan.status == 'completed') {
        timer.cancel();
        // Генерация завершена, показать план
        final cells = (response.data['data']['cells'] as List)
            .map((e) => KMJCell.fromJson(e))
            .toList();
        _showCompletedPlan(plan, cells);
      } else if (plan.status == 'failed') {
        timer.cancel();
        _showError('Генерация не удалась');
      }
      // Если status == 'generating', продолжаем ждать
    } catch (e) {
      timer.cancel();
      _showError(e.toString());
    }
  });
}

@override
void dispose() {
  _pollTimer?.cancel();
  super.dispose();
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "message": "Generation started successfully",
    "plan_id": "uuid-here"
  }
}
```

**Error Response (403):**
```json
{
  "success": false,
  "error": {
    "code": "INSUFFICIENT_TOKENS",
    "message": "insufficient tokens: ...",
    "details": "..."
  }
}
```

---

## 📖 3. Get Plan by ID

**Endpoint:** `GET /api/v1/kmj/{id}`

**Request:**
```dart
final response = await dio.get('/kmj/$planId');
final planWithCells = KMJPlanWithCells.fromJson(response.data['data']);

// Проверка статуса
switch (planWithCells.plan.status) {
  case 'draft':
    print('Черновик - нужно запустить генерацию');
    break;
  case 'generating':
    print('Генерируется - показать loader');
    break;
  case 'completed':
    print('Готово - показать ячейки');
    _displayCells(planWithCells.cells);
    break;
  case 'failed':
    print('Ошибка генерации');
    break;
}
```

**Группировка ячеек по секциям:**
```dart
Map<String, List<KMJCell>> groupCellsBySection(List<KMJCell> cells) {
  final grouped = <String, List<KMJCell>>{};
  for (var cell in cells) {
    grouped.putIfAbsent(cell.section, () => []).add(cell);
  }
  return grouped;
}

// Использование
final grouped = groupCellsBySection(planWithCells.cells);
final headerCells = grouped['header'] ?? [];
final startCells = grouped['lesson_start'] ?? [];
final middleCells = grouped['lesson_middle'] ?? [];
final endCells = grouped['lesson_end'] ?? [];
```

---

## 📝 4. Update Cell

**Endpoint:** `PUT /api/v1/kmj/cells/{cellId}`

**Request:**
```dart
Future<void> updateCell(String cellId, String newValue) async {
  try {
    final response = await dio.put(
      '/kmj/cells/$cellId',
      data: {'cell_value': newValue},
    );
    
    if (response.data['success']) {
      print('Ячейка обновлена');
      // После обновления флаг is_edited станет true
    }
  } on DioException catch (e) {
    print('Ошибка: ${e.response?.data['error']['message']}');
  }
}
```

**UI Example:**
```dart
class EditableCellWidget extends StatefulWidget {
  final KMJCell cell;
  final Function(String) onUpdate;
  
  @override
  _EditableCellWidgetState createState() => _EditableCellWidgetState();
}

class _EditableCellWidgetState extends State<EditableCellWidget> {
  late TextEditingController _controller;
  bool _isEditing = false;
  
  @override
  void initState() {
    super.initState();
    _controller = TextEditingController(text: widget.cell.cellValue);
  }
  
  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Text(widget.cell.cellKey),
            if (widget.cell.isAiGenerated)
              Chip(label: Text('AI'), backgroundColor: Colors.blue.shade100),
            if (widget.cell.isEdited)
              Icon(Icons.edit, size: 16, color: Colors.orange),
          ],
        ),
        TextField(
          controller: _controller,
          maxLines: null,
          onChanged: (_) => setState(() => _isEditing = true),
        ),
        if (_isEditing)
          ElevatedButton(
            onPressed: () {
              widget.onUpdate(_controller.text);
              setState(() => _isEditing = false);
            },
            child: Text('Сохранить'),
          ),
      ],
    );
  }
}
```

---

## 📚 5. List Plans

**Endpoint:** `GET /api/v1/kmj?page=1&page_size=50`

**Request:**
```dart
Future<List<KMJPlan>> fetchPlans({int page = 1, int pageSize = 50}) async {
  final response = await dio.get('/kmj', queryParameters: {
    'page': page,
    'page_size': pageSize,
  });
  
  final data = response.data['data'];
  final plans = (data['plans'] as List)
      .map((e) => KMJPlan.fromJson(e))
      .toList();
  
  final total = data['total'] as int;
  final totalPages = data['total_pages'] as int;
  
  return plans;
}
```

**Pagination Widget:**
```dart
class KMJPlanList extends StatefulWidget {
  @override
  _KMJPlanListState createState() => _KMJPlanListState();
}

class _KMJPlanListState extends State<KMJPlanList> {
  final ScrollController _scrollController = ScrollController();
  List<KMJPlan> _plans = [];
  int _currentPage = 1;
  bool _isLoading = false;
  bool _hasMore = true;
  
  @override
  void initState() {
    super.initState();
    _loadPlans();
    _scrollController.addListener(_onScroll);
  }
  
  void _onScroll() {
    if (_scrollController.position.pixels ==
        _scrollController.position.maxScrollExtent) {
      _loadMore();
    }
  }
  
  Future<void> _loadPlans() async {
    setState(() => _isLoading = true);
    final newPlans = await fetchPlans(page: 1);
    setState(() {
      _plans = newPlans;
      _currentPage = 1;
      _hasMore = newPlans.length == 50;
      _isLoading = false;
    });
  }
  
  Future<void> _loadMore() async {
    if (!_hasMore || _isLoading) return;
    
    setState(() => _isLoading = true);
    final newPlans = await fetchPlans(page: _currentPage + 1);
    setState(() {
      _plans.addAll(newPlans);
      _currentPage++;
      _hasMore = new Plans.length == 50;
      _isLoading = false;
    });
  }
  
  @override
  Widget build(BuildContext context) {
    return RefreshIndicator(
      onRefresh: _loadPlans,
      child: ListView.builder(
        controller: _scrollController,
        itemCount: _plans.length + (_hasMore ? 1 : 0),
        itemBuilder: (context, index) {
          if (index == _plans.length) {
            return Center(child: CircularProgressIndicator());
          }
          
          final plan = _plans[index];
          return PlanListTile(plan: plan);
        },
      ),
    );
  }
}
```

---

## 🗑️ 6. Delete Plan

**Endpoint:** `DELETE /api/v1/kmj/{id}`

**Request:**
```dart
Future<bool> deletePlan(String planId) async {
  try {
    final response = await dio.delete('/kmj/$planId');
    return response.data['success'] == true;
  } catch (e) {
    print('Delete error: $e');
    return false;
  }
}

// With confirmation dialog
Future<void> deletePlanWithConfirmation(String planId) async {
  final confirmed = await showDialog<bool>(
    context: context,
    builder: (context) => AlertDialog(
      title: Text('Удалить план?'),
      content: Text('Это действие нельзя отменить'),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context, false),
          child: Text('Отмена'),
        ),
        TextButton(
          onPressed: () => Navigator.pop(context, true),
          child: Text('Удалить'),
          style: TextButton.styleFrom(foregroundColor: Colors.red),
        ),
      ],
    ),
  );
  
  if (confirmed == true) {
    final success = await deletePlan(planId);
    if (success) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('План удален')),
      );
      _refreshPlanList();
    }
  }
}
```

---

## 🎨 UI Constants

```dart
class KMJConstants {
  // Sections
  static const String sectionHeader = 'header';
  static const String sectionLessonStart = 'lesson_start';
  static const String sectionLessonMiddle = 'lesson_middle';
  static const String sectionLessonEnd = 'lesson_end';
  
  // Stage columns
  static const String columnPedagogAction = 'pedagog_action';
  static const String columnStudentAction = 'student_action';
  static const String columnAssessment = 'assessment';
  static const String columnResources = 'resources';
  
  // Status
  static const String statusDraft = 'draft';
  static const String statusGenerating = 'generating';
  static const String statusCompleted = 'completed';
  static const String statusFailed = 'failed';
  
  // Token cost
  static const int tokenCostGeneration = 5;
}
```

---

## ⚠️ Error Handling

```dart
class KMJApiException implements Exception {
  final String code;
  final String message;
  final dynamic details;
  
  KMJApiException(this.code, this.message, [this.details]);
  
  factory KMJApiException.fromResponse(Response response) {
    final error = response.data['error'];
    return KMJApiException(
      error['code'] ?? 'UNKNOWN_ERROR',
      error['message'] ?? 'Unknown error occurred',
      error['details'],
    );
  }
  
  @override
  String toString() => '$code: $message';
}

// Global error handler
Future<T> handleApiCall<T>(Future<T> Function() apiCall) async {
  try {
    return await apiCall();
  } on DioException catch (e) {
    if (e.response != null) {
      throw KMJApiException.fromResponse(e.response!);
    } else {
      throw KMJApiException('NETWORK_ERROR', 'Network error: ${e.message}');
    }
  }
}

// Usage
try {
  final plans = await handleApiCall(() => fetchPlans());
} on KMJApiException catch (e) {
  if (e.code == 'INSUFFICIENT_TOKENS') {
    _showUpgradeDialog();
  } else {
    _showErrorSnackbar(e.message);
  }
}
```

---

## 🔄 Complete Service Class

```dart
class KMJService {
  final Dio _dio;
  
  KMJService(this._dio);
  
  Future<KMJPlanWithCells> createDraft({
    required String subject,
    required int grade,
    required int lessonNumber,
    required String lessonTopic,
    required String teacherName,
    required String contentLang,
    DateTime? lessonDate,
  }) async {
    final response = await _dio.post('/kmj', data: {
      'subject': subject,
      'grade': grade,
      'lesson_number': lessonNumber,
      'lesson_topic': lessonTopic,
      'teacher_name': teacherName,
      'content_lang': contentLang,
      if (lessonDate != null) 'lesson_date': lessonDate.toIso8601String(),
    });
    
    return KMJPlanWithCells.fromJson(response.data['data']);
  }
  
  Future<void> generate(String planId) async {
    await _dio.post('/kmj/$planId/generate');
  }
  
  Future<KMJPlanWithCells> getPlan(String planId) async {
    final response = await _dio.get('/kmj/$planId');
    return KMJPlanWithCells.fromJson(response.data['data']);
  }
  
  Future<List<KMJPlan>> listPlans({int page = 1, int pageSize = 50}) async {
    final response = await _dio.get('/kmj', queryParameters: {
      'page': page,
      'page_size': pageSize,
    });
    
    return (response.data['data']['plans'] as List)
        .map((e) => KMJPlan.fromJson(e))
        .toList();
  }
  
  Future<void> updateCell(String cellId, String value) async {
    await _dio.put('/kmj/cells/$cellId', data: {'cell_value': value});
  }
  
  Future<void> deletePlan(String planId) async {
    await _dio.delete('/kmj/$planId');
  }
}
```

---

## 📱 Complete Example Screen

```dart
class KMJDetailScreen extends StatefulWidget {
  final String planId;
  
  const KMJDetailScreen({required this.planId});
  
  @override
  _KMJDetailScreenState createState() => _KMJDetailScreenState();
}

class _KMJDetailScreenState extends State<KMJDetailScreen> {
  KMJPlanWithCells? _planData;
  bool _isLoading = true;
  Timer? _pollTimer;
  
  @override
  void initState() {
    super.initState();
    _loadPlan();
  }
  
  Future<void> _loadPlan() async {
    try {
      final service = context.read<KMJService>();
      final data = await service.getPlan(widget.planId);
      
      setState(() {
        _planData = data;
        _isLoading = false;
      });
      
      // Start polling if generating
      if (data.plan.status == KMJConstants.statusGenerating) {
        _startPolling();
      }
    } catch (e) {
      setState(() => _isLoading = false);
      _showError(e.toString());
    }
  }
  
  void _startPolling() {
    _pollTimer?.cancel();
    _pollTimer = Timer.periodic(Duration(seconds: 3), (_) => _loadPlan());
  }
  
  @override
  void dispose() {
    _pollTimer?.cancel();
    super.dispose();
  }
  
  @override
  Widget build(BuildContext context) {
    if (_isLoading) {
      return Scaffold(
        appBar: AppBar(title: Text('Загрузка...')),
        body: Center(child: CircularProgressIndicator()),
      );
    }
    
    final plan = _planData!.plan;
    final cells = _planData!.cells;
    
    return Scaffold(
      appBar: AppBar(
        title: Text(plan.lessonTopic),
        actions: [
          if (plan.status == KMJConstants.statusDraft)
            IconButton(
              icon: Icon(Icons.auto_awesome),
              tooltip: 'Генерировать (5 токенов)',
              onPressed: _generatePlan,
            ),
        ],
      ),
      body: _buildBody(plan, cells),
    );
  }
  
  Widget _buildBody(KMJPlan plan, List<KMJCell> cells) {
    switch (plan.status) {
      case KMJConstants.statusDraft:
        return _buildDraftView(cells);
      case KMJConstants.statusGenerating:
        return _buildGeneratingView();
      case KMJConstants.statusCompleted:
        return _buildCompletedView(cells);
      case KMJConstants.statusFailed:
        return _buildFailedView();
      default:
        return Center(child: Text('Unknown status'));
    }
  }
  
  Widget _buildGeneratingView() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          CircularProgressIndicator(),
          SizedBox(height: 16),
          Text('Генерация плана...'),
          Text('Это может занять до 30 секунд'),
        ],
      ),
    );
  }
  
  Future<void> _generatePlan() async {
    try {
      final service = context.read<KMJService>();
      await service.generate(widget.planId);
      _loadPlan(); // Reload to show generating status
    } on KMJApiException catch (e) {
      if (e.code == 'INSUFFICIENT_TOKENS') {
        _showUpgradeDialog();
      } else {
        _showError(e.message);
      }
    }
  }
}
```

---

## 🚀 Ready to Use!

Check Swagger UI for detailed API documentation:
- **Development:** http://localhost:8000/swagger/index.html
- **API Spec:** http://localhost:8000/swagger/doc.json

All models, request/response examples, and error codes are documented there!
