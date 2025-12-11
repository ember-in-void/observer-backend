# Твой первый ML код! 
from sklearn.datasets import load_iris
from sklearn.model_selection import train_test_split
from sklearn.tree import DecisionTreeClassifier
from sklearn.metrics import accuracy_score

# Загрузи датасет (цветы ириса)
iris = load_iris()
X, y = iris.data, iris. target

# Раздели на train/test (как в Go тестах)
X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2)

# Создай и обучи модель
model = DecisionTreeClassifier()
model.fit(X_train, y_train)

# Предскажи и оцени
predictions = model.predict(X_test)
accuracy = accuracy_score(y_test, predictions)

print(f"🎯 Точность модели: {accuracy * 100:.2f}%")
print(f"📊 Предсказано {len(predictions)} образцов")